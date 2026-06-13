package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// withAuthTestConfig swaps loadedConfig and the session secret for the test
// and restores them afterwards.
func withAuthTestConfig(t *testing.T, cfg appConfig, secret []byte) {
	t.Helper()
	oldCfg := loadedConfig
	oldSecret := currentSessionSecret()
	loadedConfig = cfg
	setSessionSecret(secret)
	t.Cleanup(func() {
		loadedConfig = oldCfg
		setSessionSecret(oldSecret)
	})
}

func authTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	handleAPI(mux, "GET /api/v1/players", capPlayersRead, func(w http.ResponseWriter, r *http.Request) {
		jsonOK(w, map[string]string{"ok": "read"})
	})
	handleAPI(mux, "POST /api/v1/players/give-item", capPlayersWrite, func(w http.ResponseWriter, r *http.Request) {
		jsonOK(w, map[string]string{"ok": "write"})
	})
	mux.HandleFunc("GET /api/v1/auth/status", func(w http.ResponseWriter, r *http.Request) {
		jsonOK(w, map[string]bool{"auth": true})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "spa")
	})
	return mux
}

func mintTestSession(t *testing.T, claims sessionClaims, secret []byte) string {
	t.Helper()
	tok, err := mintSession(claims, secret, time.Hour)
	if err != nil {
		t.Fatalf("mintSession: %v", err)
	}
	return tok
}

func doAuthRequest(handler http.Handler, method, path, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, nil)
	r.RemoteAddr = "203.0.113.10:5555" // non-loopback public client
	if token != "" {
		r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestAuthMiddlewareDisabled(t *testing.T) {
	withAuthTestConfig(t, appConfig{}, []byte("0123456789abcdef0123456789abcdef"))
	mux := authTestMux(t)
	handler := authMiddleware(mux, mux)

	// Everything passes through untouched, no cookie needed.
	for _, tc := range []struct{ method, path string }{
		{"GET", "/api/v1/players"},
		{"POST", "/api/v1/players/give-item"},
		{"GET", "/api/v1/auth/status"},
		{"GET", "/index.html"},
	} {
		w := doAuthRequest(handler, tc.method, tc.path, "")
		if w.Code != http.StatusOK {
			t.Errorf("%s %s: status %d, want 200 (auth disabled)", tc.method, tc.path, w.Code)
		}
	}
}

func TestAuthMiddlewareEnabled(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	enabled := true
	cfg := appConfig{AuthEnabled: &enabled}
	withAuthTestConfig(t, cfg, secret)
	mux := authTestMux(t)
	handler := authMiddleware(mux, mux)

	ownerTok := mintTestSession(t, sessionClaims{Sub: "local:admin", Method: "local", Owner: true}, secret)
	readerTok := mintTestSession(t, sessionClaims{
		Sub: "discord:42", Method: "discord", RoleIDs: []string{"999"},
		RolesAt: time.Now().Unix(),
	}, secret)
	badTok := "garbage.token.here"

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		wantStatus int
	}{
		{"no cookie API read → 401", "GET", "/api/v1/players", "", http.StatusUnauthorized},
		{"no cookie API write → 401", "POST", "/api/v1/players/give-item", "", http.StatusUnauthorized},
		{"invalid cookie → 401", "GET", "/api/v1/players", badTok, http.StatusUnauthorized},
		{"auth endpoints exempt", "GET", "/api/v1/auth/status", "", http.StatusOK},
		{"SPA assets exempt", "GET", "/index.html", "", http.StatusOK},
		{"owner can read", "GET", "/api/v1/players", ownerTok, http.StatusOK},
		{"owner can write", "POST", "/api/v1/players/give-item", ownerTok, http.StatusOK},
		{"default member can read", "GET", "/api/v1/players", readerTok, http.StatusOK},
		{"default member cannot write → 403", "POST", "/api/v1/players/give-item", readerTok, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doAuthRequest(handler, tt.method, tt.path, tt.token)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}

	t.Run("unmapped API route fails closed → 403", func(t *testing.T) {
		// Registered directly on the mux without handleAPI — no capability.
		mux.HandleFunc("GET /api/v1/sneaky", func(w http.ResponseWriter, r *http.Request) {
			jsonOK(w, "should never run")
		})
		w := doAuthRequest(handler, "GET", "/api/v1/sneaky", ownerTok)
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403 for unmapped route", w.Code)
		}
	})

	t.Run("cross-origin mutation blocked", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/players/give-item", nil)
		r.RemoteAddr = "203.0.113.10:5555"
		r.Header.Set("Origin", "https://evil.example.com")
		r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ownerTok})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403 for cross-origin mutation", w.Code)
		}
	})

	t.Run("same-origin mutation allowed", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "http://dash.example.com/api/v1/players/give-item", nil)
		r.Host = "dash.example.com"
		r.RemoteAddr = "203.0.113.10:5555"
		r.Header.Set("Origin", "http://dash.example.com")
		r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ownerTok})
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 for same-origin mutation (body: %s)", w.Code, w.Body.String())
		}
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("headers absent when auth disabled", func(t *testing.T) {
		withAuthTestConfig(t, appConfig{}, nil)
		w := httptest.NewRecorder()
		securityHeadersMiddleware(inner).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if got := w.Header().Get("X-Frame-Options"); got != "" {
			t.Errorf("X-Frame-Options = %q, want unset when auth disabled", got)
		}
	})

	t.Run("headers present when auth enabled", func(t *testing.T) {
		enabled := true
		withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, nil)
		w := httptest.NewRecorder()
		securityHeadersMiddleware(inner).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("X-Content-Type-Options = %q", got)
		}
		if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
			t.Errorf("X-Frame-Options = %q", got)
		}
		if got := w.Header().Get("Referrer-Policy"); got == "" {
			t.Error("Referrer-Policy unset")
		}
		if got := w.Header().Get("Content-Security-Policy"); got == "" {
			t.Error("Content-Security-Policy unset")
		}
		if got := w.Header().Get("Strict-Transport-Security"); got != "" {
			t.Errorf("HSTS = %q, want unset over plain HTTP", got)
		}
	})

	t.Run("HSTS only behind TLS", func(t *testing.T) {
		enabled := true
		withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, nil)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("X-Forwarded-Proto", "https")
		w := httptest.NewRecorder()
		securityHeadersMiddleware(inner).ServeHTTP(w, r)
		if got := w.Header().Get("Strict-Transport-Security"); got == "" {
			t.Error("HSTS unset behind TLS proxy")
		}
	})

	t.Run("swagger gets relaxed CSP allowing inline bootstrap", func(t *testing.T) {
		enabled := true
		withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, nil)
		w := httptest.NewRecorder()
		securityHeadersMiddleware(inner).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))
		csp := w.Header().Get("Content-Security-Policy")
		if !strings.Contains(csp, "script-src 'self' 'unsafe-inline'") {
			t.Errorf("swagger CSP missing inline script allowance: %q", csp)
		}
	})

	t.Run("app CSP forbids inline scripts", func(t *testing.T) {
		enabled := true
		withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, nil)
		w := httptest.NewRecorder()
		securityHeadersMiddleware(inner).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		csp := w.Header().Get("Content-Security-Policy")
		if strings.Contains(csp, "script-src") && strings.Contains(csp, "'unsafe-inline'") {
			// style-src 'unsafe-inline' is fine; ensure script-src does NOT have it
			if strings.Contains(csp, "script-src 'self' 'unsafe-inline'") {
				t.Errorf("app CSP must not allow inline scripts: %q", csp)
			}
		}
	})

	// The LiveMap loads tiles/images from external CDNs (cdn.th.gl, the
	// configurable VITE_CDN_BASE_URL host). img-src must allow https: or those
	// images are blocked once auth (and thus CSP) is enabled (#livemap).
	t.Run("app CSP allows external https images", func(t *testing.T) {
		enabled := true
		withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, nil)
		w := httptest.NewRecorder()
		securityHeadersMiddleware(inner).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		csp := w.Header().Get("Content-Security-Policy")
		if !strings.Contains(csp, "img-src 'self' data: https:") {
			t.Errorf("app CSP must allow https images for map tiles: %q", csp)
		}
	})
}

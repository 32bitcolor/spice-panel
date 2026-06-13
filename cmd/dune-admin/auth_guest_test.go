package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func guestEnabledConfig(t *testing.T) appConfig {
	t.Helper()
	cfg := localAuthConfig(t, "pw")
	enabled := true
	cfg.AuthGuestEnabled = &enabled
	return cfg
}

func TestHandleAuthGuest(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")

	t.Run("mints read-only guest session when enabled", func(t *testing.T) {
		withAuthTestConfig(t, guestEnabledConfig(t), secret)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/guest", nil)
		w := httptest.NewRecorder()
		handleAuthGuest(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != sessionCookieName {
			t.Fatalf("expected session cookie, got %v", cookies)
		}
		claims, err := verifySession(cookies[0].Value, secret)
		if err != nil {
			t.Fatal(err)
		}
		if claims.Method != "guest" || claims.Owner {
			t.Errorf("claims = %+v", claims)
		}
	})

	t.Run("guest session passes read routes and fails writes", func(t *testing.T) {
		withAuthTestConfig(t, guestEnabledConfig(t), secret)
		mux := authTestMux(t)
		handler := authMiddleware(mux, mux)
		tok := mintTestSession(t, sessionClaims{Sub: "guest", Name: "Guest", Method: "guest"}, secret)
		if w := doAuthRequest(handler, "GET", "/api/v1/players", tok); w.Code != http.StatusOK {
			t.Errorf("guest read = %d, want 200", w.Code)
		}
		if w := doAuthRequest(handler, "POST", "/api/v1/players/give-item", tok); w.Code != http.StatusForbidden {
			t.Errorf("guest write = %d, want 403", w.Code)
		}
	})

	t.Run("404 when guest access disabled", func(t *testing.T) {
		withAuthTestConfig(t, localAuthConfig(t, "pw"), secret)
		w := httptest.NewRecorder()
		handleAuthGuest(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth/guest", nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("404 when auth disabled entirely", func(t *testing.T) {
		withAuthTestConfig(t, appConfig{}, secret)
		w := httptest.NewRecorder()
		handleAuthGuest(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth/guest", nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("status advertises guest method", func(t *testing.T) {
		withAuthTestConfig(t, guestEnabledConfig(t), secret)
		w := httptest.NewRecorder()
		handleAuthStatus(w, httptest.NewRequest(http.MethodGet, "/api/v1/auth/status", nil))
		var resp struct {
			Methods struct {
				Guest bool `json:"guest"`
			} `json:"methods"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if !resp.Methods.Guest {
			t.Error("guest method not advertised")
		}
	})
}

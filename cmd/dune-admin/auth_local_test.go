package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func localAuthConfig(t *testing.T, password string) appConfig {
	t.Helper()
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	enabled := true
	return appConfig{
		AuthEnabled:           &enabled,
		AuthLocalUsername:     "admin",
		AuthLocalPasswordHash: hash,
	}
}

func postLogin(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.RemoteAddr = "203.0.113.20:1234"
	w := httptest.NewRecorder()
	handleAuthLogin(w, r)
	return w
}

func TestHandleAuthLogin(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")

	t.Run("valid credentials set owner session cookie", func(t *testing.T) {
		withAuthTestConfig(t, localAuthConfig(t, "hunter22"), secret)
		resetLoginLimiter()
		w := postLogin(t, `{"username":"admin","password":"hunter22"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body %s", w.Code, w.Body.String())
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != sessionCookieName {
			t.Fatalf("expected session cookie, got %v", cookies)
		}
		claims, err := verifySession(cookies[0].Value, secret)
		if err != nil {
			t.Fatalf("cookie token invalid: %v", err)
		}
		if !claims.Owner {
			t.Error("local account must be owner")
		}
		if claims.Method != "local" || claims.Sub != "local:admin" {
			t.Errorf("claims = %+v", claims)
		}
	})

	t.Run("wrong password → 401, no cookie", func(t *testing.T) {
		withAuthTestConfig(t, localAuthConfig(t, "hunter22"), secret)
		resetLoginLimiter()
		w := postLogin(t, `{"username":"admin","password":"wrong"}`)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
		if len(w.Result().Cookies()) != 0 {
			t.Error("cookie set on failed login")
		}
	})

	t.Run("wrong username → 401", func(t *testing.T) {
		withAuthTestConfig(t, localAuthConfig(t, "hunter22"), secret)
		resetLoginLimiter()
		w := postLogin(t, `{"username":"someoneelse","password":"hunter22"}`)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})

	t.Run("local login unconfigured → 401", func(t *testing.T) {
		enabled := true
		withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, secret)
		resetLoginLimiter()
		w := postLogin(t, `{"username":"admin","password":"anything"}`)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})

	t.Run("malformed body → 400", func(t *testing.T) {
		withAuthTestConfig(t, localAuthConfig(t, "hunter22"), secret)
		resetLoginLimiter()
		w := postLogin(t, `{nope`)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("auth disabled → 404", func(t *testing.T) {
		withAuthTestConfig(t, appConfig{}, secret)
		resetLoginLimiter()
		w := postLogin(t, `{"username":"admin","password":"x"}`)
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404 when auth disabled", w.Code)
		}
	})
}

func TestHandleAuthLogout(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	withAuthTestConfig(t, localAuthConfig(t, "pw"), secret)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	handleAuthLogout(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge != -1 {
		t.Errorf("expected expiring cookie, got %v", cookies)
	}
}

func TestHandleAuthStatus(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")

	t.Run("disabled reports enabled=false", func(t *testing.T) {
		withAuthTestConfig(t, appConfig{}, secret)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/status", nil)
		w := httptest.NewRecorder()
		handleAuthStatus(w, r)
		var resp authStatusResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Enabled {
			t.Error("enabled = true, want false")
		}
	})

	t.Run("enabled without session reports methods", func(t *testing.T) {
		withAuthTestConfig(t, localAuthConfig(t, "pw"), secret)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/status", nil)
		w := httptest.NewRecorder()
		handleAuthStatus(w, r)
		var resp authStatusResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if !resp.Enabled || !resp.Methods.Local || resp.Methods.Discord {
			t.Errorf("resp = %+v", resp)
		}
		if resp.Session != nil {
			t.Error("session reported without cookie")
		}
	})

	t.Run("enabled with session reports identity and caps", func(t *testing.T) {
		withAuthTestConfig(t, localAuthConfig(t, "pw"), secret)
		tok := mintTestSession(t, sessionClaims{Sub: "local:admin", Name: "admin", Method: "local", Owner: true}, secret)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/status", nil)
		r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tok})
		w := httptest.NewRecorder()
		handleAuthStatus(w, r)
		var resp authStatusResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Session == nil {
			t.Fatal("session missing")
		}
		if !resp.Session.Owner || resp.Session.Sub != "local:admin" {
			t.Errorf("session = %+v", resp.Session)
		}
		if len(resp.Session.Capabilities) != len(allCapabilities) {
			t.Errorf("owner capabilities = %d, want all %d", len(resp.Session.Capabilities), len(allCapabilities))
		}
	})
}

func TestLoginRateLimiter(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	clock := func() time.Time { return now }

	t.Run("allows up to limit then blocks", func(t *testing.T) {
		lim := newLoginRateLimiter(clock)
		for i := range loginRateLimit {
			if !lim.allow("1.2.3.4", "admin") {
				t.Fatalf("attempt %d blocked, want allowed", i+1)
			}
		}
		if lim.allow("1.2.3.4", "admin") {
			t.Error("attempt over limit allowed")
		}
	})

	t.Run("window slides", func(t *testing.T) {
		lim := newLoginRateLimiter(clock)
		for range loginRateLimit {
			lim.allow("1.2.3.4", "admin")
		}
		now = now.Add(61 * time.Second)
		defer func() { now = time.Unix(1_700_000_000, 0) }()
		if !lim.allow("1.2.3.4", "admin") {
			t.Error("blocked after window expired")
		}
	})

	t.Run("limits are per key", func(t *testing.T) {
		lim := newLoginRateLimiter(clock)
		for range loginRateLimit {
			lim.allow("1.2.3.4", "admin")
		}
		if !lim.allow("5.6.7.8", "admin") {
			t.Error("different IP blocked")
		}
	})

	t.Run("username dimension blocks across IPs", func(t *testing.T) {
		lim := newLoginRateLimiter(clock)
		// Exhaust the username budget from many IPs.
		for i := 0; i < loginUsernameRateLimit; i++ {
			lim.allow("10.0.0."+string(rune('1'+i%9)), "victim")
		}
		if lim.allow("99.99.99.99", "victim") {
			t.Error("username brute force across IPs not blocked")
		}
	})

	t.Run("rate limited login returns 429", func(t *testing.T) {
		secret := []byte("0123456789abcdef0123456789abcdef")
		withAuthTestConfig(t, localAuthConfig(t, "pw"), secret)
		resetLoginLimiter()
		for range loginRateLimit {
			postLogin(t, `{"username":"admin","password":"wrong"}`)
		}
		w := postLogin(t, `{"username":"admin","password":"pw"}`)
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("status = %d, want 429", w.Code)
		}
	})
}

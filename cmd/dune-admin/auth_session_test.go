package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestLoadOrCreateSessionSecret(t *testing.T) {
	t.Run("creates secret on first call", func(t *testing.T) {
		dir := t.TempDir()
		secret, err := loadOrCreateSessionSecret(dir)
		if err != nil {
			t.Fatalf("loadOrCreateSessionSecret: %v", err)
		}
		if len(secret) != 32 {
			t.Fatalf("secret length = %d, want 32", len(secret))
		}
		path := filepath.Join(dir, "session-secret")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
			t.Errorf("file mode = %v, want 0600", info.Mode().Perm())
		}
	})

	t.Run("returns same secret on subsequent calls", func(t *testing.T) {
		dir := t.TempDir()
		first, err := loadOrCreateSessionSecret(dir)
		if err != nil {
			t.Fatalf("first call: %v", err)
		}
		second, err := loadOrCreateSessionSecret(dir)
		if err != nil {
			t.Fatalf("second call: %v", err)
		}
		if string(first) != string(second) {
			t.Error("secrets differ between calls")
		}
	})

	t.Run("different dirs get different secrets", func(t *testing.T) {
		a, err := loadOrCreateSessionSecret(t.TempDir())
		if err != nil {
			t.Fatalf("a: %v", err)
		}
		b, err := loadOrCreateSessionSecret(t.TempDir())
		if err != nil {
			t.Fatalf("b: %v", err)
		}
		if string(a) == string(b) {
			t.Error("two installs produced identical secrets")
		}
	})

	t.Run("unwritable dir returns error", func(t *testing.T) {
		if runtime.GOOS == "windows" || os.Geteuid() == 0 {
			t.Skip("permission test not reliable here")
		}
		dir := filepath.Join(t.TempDir(), "ro")
		if err := os.Mkdir(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		if _, err := loadOrCreateSessionSecret(dir); err == nil {
			t.Error("expected error for unwritable dir")
		}
	})
}

func TestMintAndVerifySession(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	otherSecret := []byte("fedcba9876543210fedcba9876543210")
	baseClaims := sessionClaims{
		Sub:     "discord:123456",
		Name:    "TestUser",
		Method:  "discord",
		RoleIDs: []string{"111", "222"},
		Owner:   true,
		RolesAt: time.Now().Unix(),
	}

	t.Run("round trip preserves claims", func(t *testing.T) {
		tok, err := mintSession(baseClaims, secret, time.Hour)
		if err != nil {
			t.Fatalf("mintSession: %v", err)
		}
		got, err := verifySession(tok, secret)
		if err != nil {
			t.Fatalf("verifySession: %v", err)
		}
		if got.Sub != baseClaims.Sub || got.Name != baseClaims.Name ||
			got.Method != baseClaims.Method || got.Owner != baseClaims.Owner ||
			got.RolesAt != baseClaims.RolesAt {
			t.Errorf("claims mismatch: got %+v, want %+v", got, baseClaims)
		}
		if len(got.RoleIDs) != 2 || got.RoleIDs[0] != "111" || got.RoleIDs[1] != "222" {
			t.Errorf("RoleIDs = %v, want [111 222]", got.RoleIDs)
		}
	})

	t.Run("expired token rejected", func(t *testing.T) {
		tok, err := mintSession(baseClaims, secret, -time.Minute)
		if err != nil {
			t.Fatalf("mintSession: %v", err)
		}
		if _, err := verifySession(tok, secret); err == nil {
			t.Error("expired token verified")
		}
	})

	t.Run("wrong secret rejected", func(t *testing.T) {
		tok, err := mintSession(baseClaims, secret, time.Hour)
		if err != nil {
			t.Fatalf("mintSession: %v", err)
		}
		if _, err := verifySession(tok, otherSecret); err == nil {
			t.Error("token signed with different secret verified")
		}
	})

	t.Run("tampered payload rejected", func(t *testing.T) {
		tok, err := mintSession(baseClaims, secret, time.Hour)
		if err != nil {
			t.Fatalf("mintSession: %v", err)
		}
		// Flip a byte mid-token (payload region).
		b := []byte(tok)
		mid := len(b) / 2
		if b[mid] == 'A' {
			b[mid] = 'B'
		} else {
			b[mid] = 'A'
		}
		if _, err := verifySession(string(b), secret); err == nil {
			t.Error("tampered token verified")
		}
	})

	t.Run("garbage rejected", func(t *testing.T) {
		for _, tok := range []string{"", "not-a-jwt", "a.b.c"} {
			if _, err := verifySession(tok, secret); err == nil {
				t.Errorf("garbage token %q verified", tok)
			}
		}
	})

	t.Run("alg none rejected", func(t *testing.T) {
		// header {"alg":"none","typ":"JWT"} + payload {"sub":"x"} + empty sig
		tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ4In0."
		if _, err := verifySession(tok, secret); err == nil {
			t.Error("alg=none token verified")
		}
	})
}

func TestSessionCookie(t *testing.T) {
	t.Run("sets HttpOnly Lax cookie over plain HTTP", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		setSessionCookie(w, r, "tok123", time.Hour)
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("got %d cookies, want 1", len(cookies))
		}
		c := cookies[0]
		if c.Name != sessionCookieName {
			t.Errorf("name = %q, want %q", c.Name, sessionCookieName)
		}
		if c.Value != "tok123" {
			t.Errorf("value = %q", c.Value)
		}
		if !c.HttpOnly {
			t.Error("cookie not HttpOnly")
		}
		if c.SameSite != http.SameSiteLaxMode {
			t.Errorf("SameSite = %v, want Lax", c.SameSite)
		}
		if c.Secure {
			t.Error("Secure set on plain-HTTP request")
		}
		if c.Path != "/" {
			t.Errorf("path = %q, want /", c.Path)
		}
	})

	t.Run("sets Secure when X-Forwarded-Proto is https", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		r.Header.Set("X-Forwarded-Proto", "https")
		setSessionCookie(w, r, "tok", time.Hour)
		if !w.Result().Cookies()[0].Secure {
			t.Error("Secure not set behind https proxy")
		}
	})

	t.Run("clearSessionCookie expires the cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		clearSessionCookie(w, r)
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("got %d cookies, want 1", len(cookies))
		}
		if cookies[0].MaxAge != -1 {
			t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
		}
		if cookies[0].Value != "" {
			t.Errorf("value = %q, want empty", cookies[0].Value)
		}
	})
}

func TestAuthEnabled(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }
	tests := []struct {
		name string
		cfg  appConfig
		want bool
	}{
		{"nil → disabled", appConfig{}, false},
		{"explicit false", appConfig{AuthEnabled: boolPtr(false)}, false},
		{"explicit true", appConfig{AuthEnabled: boolPtr(true)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := authEnabled(tt.cfg); got != tt.want {
				t.Errorf("authEnabled = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthSecretsMasking(t *testing.T) {
	t.Run("mask covers auth secrets", func(t *testing.T) {
		cfg := appConfig{
			AuthDiscordClientSecret: "supersecret",
			AuthLocalPasswordHash:   "$2a$10$hash",
		}
		maskSecrets(&cfg)
		if cfg.AuthDiscordClientSecret != masked {
			t.Errorf("client secret not masked: %q", cfg.AuthDiscordClientSecret)
		}
		if cfg.AuthLocalPasswordHash != masked {
			t.Errorf("password hash not masked: %q", cfg.AuthLocalPasswordHash)
		}
	})

	t.Run("preserve restores masked auth secrets", func(t *testing.T) {
		oldLoaded := loadedConfig
		defer func() { loadedConfig = oldLoaded }()
		loadedConfig = appConfig{
			AuthDiscordClientSecret: "real-secret",
			AuthLocalPasswordHash:   "real-hash",
		}
		cfg := appConfig{
			AuthDiscordClientSecret: masked,
			AuthLocalPasswordHash:   masked,
		}
		preserveMaskedSecrets(&cfg, func(string) ([]byte, error) {
			return nil, os.ErrNotExist
		}, "ignored")
		if cfg.AuthDiscordClientSecret != "real-secret" {
			t.Errorf("client secret = %q, want restored", cfg.AuthDiscordClientSecret)
		}
		if cfg.AuthLocalPasswordHash != "real-hash" {
			t.Errorf("password hash = %q, want restored", cfg.AuthLocalPasswordHash)
		}
	})
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := hashPassword("hunter22")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if hash == "hunter22" {
		t.Fatal("hash equals plaintext")
	}
	if !checkPassword(hash, "hunter22") {
		t.Error("correct password rejected")
	}
	if checkPassword(hash, "wrong") {
		t.Error("wrong password accepted")
	}
	if checkPassword("", "anything") {
		t.Error("empty hash accepted a password")
	}
}

func TestSessionSameSite(t *testing.T) {
	tests := []struct {
		name string
		cfg  string
		want http.SameSite
	}{
		{"empty → Lax", "", http.SameSiteLaxMode},
		{"lax", "lax", http.SameSiteLaxMode},
		{"strict", "strict", http.SameSiteStrictMode},
		{"none", "none", http.SameSiteNoneMode},
		{"case-insensitive", "Strict", http.SameSiteStrictMode},
		{"whitespace", " none ", http.SameSiteNoneMode},
		{"unknown → Lax", "weird", http.SameSiteLaxMode},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sessionSameSite(appConfig{AuthCookieSameSite: tt.cfg})
			if got != tt.want {
				t.Errorf("sessionSameSite(%q) = %v, want %v", tt.cfg, got, tt.want)
			}
		})
	}
}

func TestSessionTTL(t *testing.T) {
	tests := []struct {
		name string
		cfg  appConfig
		want time.Duration
	}{
		{"default 24h", appConfig{}, 24 * time.Hour},
		{"configured 8h", appConfig{AuthSessionTTLHours: 8}, 8 * time.Hour},
		{"negative → default", appConfig{AuthSessionTTLHours: -1}, 24 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sessionTTL(tt.cfg); got != tt.want {
				t.Errorf("sessionTTL = %v, want %v", got, tt.want)
			}
		})
	}
}

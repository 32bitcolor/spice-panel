package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "dune_admin_session"

// sessionClaims is the payload carried in the dashboard session JWT.
type sessionClaims struct {
	// Sub identifies the principal: "local:<username>" or "discord:<user-id>".
	Sub string `json:"sub"`
	// Name is the display name shown in the UI.
	Name string `json:"name"`
	// Method is "local" or "discord".
	Method string `json:"method"`
	// Avatar is the Discord CDN avatar URL (empty for local logins).
	Avatar string `json:"avatar,omitempty"`
	// RoleIDs is the Discord role snapshot taken at RolesAt (empty for local).
	RoleIDs []string `json:"role_ids,omitempty"`
	// Owner grants full capabilities, bypassing the permissions matrix.
	Owner bool `json:"owner"`
	// RolesAt is the unix time the role snapshot was fetched; the middleware
	// lazily refreshes roles when this is older than rolesMaxAge.
	RolesAt int64 `json:"roles_at,omitempty"`

	jwt.RegisteredClaims
}

// authEnabled reports whether dashboard authentication is turned on.
// Missing yaml key → off, so existing installs are unaffected.
func authEnabled(cfg appConfig) bool {
	return cfg.AuthEnabled != nil && *cfg.AuthEnabled
}

// sessionTTL returns the configured session lifetime, defaulting to 24h.
func sessionTTL(cfg appConfig) time.Duration {
	if cfg.AuthSessionTTLHours > 0 {
		return time.Duration(cfg.AuthSessionTTLHours) * time.Hour
	}
	return 24 * time.Hour
}

// loadOrCreateSessionSecret returns the per-install HMAC signing key,
// generating and persisting a new 32-byte secret on first use so sessions
// survive restarts.
func loadOrCreateSessionSecret(dir string) ([]byte, error) {
	path := filepath.Join(dir, "session-secret")
	if data, err := os.ReadFile(path); err == nil && len(data) >= 32 {
		return data[:32], nil
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("generate session secret: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, secret, 0o600); err != nil {
		return nil, fmt.Errorf("persist session secret: %w", err)
	}
	return secret, nil
}

// mintSession signs a session token containing claims, valid for ttl.
func mintSession(claims sessionClaims, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(ttl))
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign session: %w", err)
	}
	return signed, nil
}

// verifySession parses and validates a session token, returning its claims.
func verifySession(token string, secret []byte) (*sessionClaims, error) {
	var claims sessionClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, fmt.Errorf("verify session: %w", err)
	}
	return &claims, nil
}

// requestIsTLS reports whether the request arrived over HTTPS, either
// directly or via a TLS-terminating reverse proxy.
func requestIsTLS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// sessionSameSite maps the auth_cookie_samesite config value to a SameSite
// mode. Default (empty/unknown) is Lax — it allows the Discord OAuth redirect
// to carry the cookie while still blocking cross-site CSRF. "strict" is the
// hardest setting (incompatible with Discord login); "none" supports
// cross-origin CDN/split-host setups and requires TLS.
func sessionSameSite(cfg appConfig) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(cfg.AuthCookieSameSite)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl / time.Second),
		HttpOnly: true,
		Secure:   requestIsTLS(r),
		SameSite: sessionSameSite(loadedConfig),
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsTLS(r),
		SameSite: sessionSameSite(loadedConfig),
	})
}

// hashPassword bcrypts a plaintext password for storage in config.yaml.
func hashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// checkPassword verifies plain against a stored bcrypt hash. An empty hash
// never matches — auth without a configured password must fail closed.
func checkPassword(hash, plain string) bool {
	if hash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

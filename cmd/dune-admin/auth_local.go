package main

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// registerAuthRoutes registers the authentication endpoints. These are
// exempt from capability enforcement (the middleware skips /api/v1/auth/)
// so the login page can function without a session.
func registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/auth/status", handleAuthStatus)
	mux.HandleFunc("POST /api/v1/auth/login", handleAuthLogin)
	mux.HandleFunc("POST /api/v1/auth/logout", handleAuthLogout)
	mux.HandleFunc("GET /api/v1/auth/discord/login", handleAuthDiscordLogin)
	mux.HandleFunc("GET /api/v1/auth/discord/callback", handleAuthDiscordCallback)
	mux.HandleFunc("GET /api/v1/auth/permissions", handleGetPermissions)
	mux.HandleFunc("PUT /api/v1/auth/permissions", handlePutPermissions)
	mux.HandleFunc("POST /api/v1/auth/guest", handleAuthGuest)
	mux.HandleFunc("GET /api/v1/auth/users", handleListAuthUsers)
	mux.HandleFunc("PUT /api/v1/auth/users/{username}", handlePutAuthUser)
	mux.HandleFunc("DELETE /api/v1/auth/users/{username}", handleDeleteAuthUser)
}

// guestLoginEnabled reports whether anonymous read-only guest sessions are
// allowed. Off unless explicitly enabled.
func guestLoginEnabled(cfg appConfig) bool {
	return cfg.AuthGuestEnabled != nil && *cfg.AuthGuestEnabled
}

// handleAuthGuest mints an anonymous read-only session. Guests get the
// default read-only capability set — no exports, config, database, logs,
// or backups.
//
// @Summary Start a read-only guest session
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/auth/guest [post]
func handleAuthGuest(w http.ResponseWriter, r *http.Request) {
	cfg := loadedConfig
	if !authEnabled(cfg) || !guestLoginEnabled(cfg) {
		jsonErr(w, errors.New("guest access is not enabled"), http.StatusNotFound)
		return
	}
	secret := currentSessionSecret()
	if secret == nil {
		jsonErr(w, errors.New("session secret not initialized"), http.StatusInternalServerError)
		return
	}
	ttl := sessionTTL(cfg)
	token, err := mintSession(sessionClaims{
		Sub:    "guest",
		Name:   "Guest",
		Method: "guest",
	}, secret, ttl)
	if err != nil {
		jsonErr(w, errors.New("could not create session"), http.StatusInternalServerError)
		return
	}
	setSessionCookie(w, r, token, ttl)
	jsonOK(w, map[string]string{"status": "ok"})
}

// ── login rate limiting ───────────────────────────────────────────────────

const (
	// loginRateLimit caps attempts per IP per window.
	loginRateLimit = 5
	// loginUsernameRateLimit caps attempts per username per window across
	// all IPs, so distributed brute force against one account still stalls.
	loginUsernameRateLimit = 10
	loginRateWindow        = time.Minute
)

type loginRateLimiter struct {
	mu       sync.Mutex
	now      func() time.Time
	attempts map[string][]time.Time
}

func newLoginRateLimiter(now func() time.Time) *loginRateLimiter {
	return &loginRateLimiter{now: now, attempts: map[string][]time.Time{}}
}

// allow records a login attempt and reports whether it is within both the
// per-IP and per-username budgets.
func (l *loginRateLimiter) allow(ip, username string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	ipOK := l.bump("ip:"+ip, now, loginRateLimit)
	userOK := l.bump("user:"+username, now, loginUsernameRateLimit)
	return ipOK && userOK
}

// bump prunes expired attempts for key, appends the current one, and reports
// whether the count is within limit. Callers hold l.mu.
func (l *loginRateLimiter) bump(key string, now time.Time, limit int) bool {
	kept := l.attempts[key][:0]
	for _, t := range l.attempts[key] {
		if now.Sub(t) < loginRateWindow {
			kept = append(kept, t)
		}
	}
	kept = append(kept, now)
	l.attempts[key] = kept
	return len(kept) <= limit
}

var globalLoginLimiter = newLoginRateLimiter(time.Now)

// resetLoginLimiter clears rate-limit state (test helper).
func resetLoginLimiter() {
	globalLoginLimiter = newLoginRateLimiter(time.Now)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ── login / logout handlers ───────────────────────────────────────────────

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// handleAuthLogin authenticates the local username/password and mints a
// session cookie. The local account is always an owner.
//
// @Summary Local username/password login
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body loginRequest true "Login credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Router /api/v1/auth/login [post]
func handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	cfg := loadedConfig
	if !authEnabled(cfg) {
		jsonErr(w, errors.New("authentication is not enabled"), http.StatusNotFound)
		return
	}
	var req loginRequest
	if err := decode(r, &req); err != nil {
		jsonErr(w, fmt.Errorf("decode: %w", err), http.StatusBadRequest)
		return
	}
	if !globalLoginLimiter.allow(clientIP(r), req.Username) {
		jsonErr(w, errors.New("too many login attempts — try again in a minute"), http.StatusTooManyRequests)
		return
	}
	// Constant-time username compare + bcrypt (constant-time by design) so
	// failures don't leak which field was wrong. The config.yaml admin is
	// always an owner; DB-stored users carry their own capability lists.
	usernameOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(cfg.AuthLocalUsername)) == 1
	passwordOK := checkPassword(cfg.AuthLocalPasswordHash, req.Password)
	isConfigAdmin := cfg.AuthLocalUsername != "" && cfg.AuthLocalPasswordHash != "" && usernameOK && passwordOK
	isDBUser := false
	if !isConfigAdmin && authUsersDB != nil {
		_, isDBUser = authUsersDB.verify(req.Username, req.Password)
	}
	if !isConfigAdmin && !isDBUser {
		jsonErr(w, errors.New("invalid username or password"), http.StatusUnauthorized)
		return
	}

	secret := currentSessionSecret()
	if secret == nil {
		jsonErr(w, errors.New("session secret not initialized"), http.StatusInternalServerError)
		return
	}
	ttl := sessionTTL(cfg)
	token, err := mintSession(sessionClaims{
		Sub:    "local:" + req.Username,
		Name:   req.Username,
		Method: "local",
		Owner:  isConfigAdmin,
	}, secret, ttl)
	if err != nil {
		jsonErr(w, errors.New("could not create session"), http.StatusInternalServerError)
		return
	}
	setSessionCookie(w, r, token, ttl)
	jsonOK(w, map[string]string{"status": "ok"})
}

// handleAuthLogout clears the session cookie.
//
// @Summary Log out (clear session cookie)
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w, r)
	jsonOK(w, map[string]string{"status": "logged out"})
}

// authStatusResponse tells the SPA whether auth is on, which login methods
// are available, and who (if anyone) the current session belongs to.
type authStatusResponse struct {
	Enabled bool `json:"enabled"`
	Methods struct {
		Local   bool `json:"local"`
		Discord bool `json:"discord"`
		Guest   bool `json:"guest"`
	} `json:"methods"`
	Session *authSessionInfo `json:"session"`
}

type authSessionInfo struct {
	Sub          string   `json:"sub"`
	Name         string   `json:"name"`
	Method       string   `json:"method"`
	Avatar       string   `json:"avatar,omitempty"`
	Owner        bool     `json:"owner"`
	Capabilities []string `json:"capabilities"`
}

// handleAuthStatus reports auth availability and the current session.
//
// @Summary Auth status: enabled flag, login methods, current session
// @Tags auth
// @Produce json
// @Success 200 {object} authStatusResponse
// @Router /api/v1/auth/status [get]
func handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	cfg := loadedConfig
	var resp authStatusResponse
	resp.Enabled = authEnabled(cfg)
	if !resp.Enabled {
		jsonOK(w, resp)
		return
	}
	resp.Methods.Local = cfg.AuthLocalPasswordHash != "" && cfg.AuthLocalUsername != ""
	resp.Methods.Discord = discordLoginConfigured(cfg)
	resp.Methods.Guest = guestLoginEnabled(cfg)

	claims, err := sessionFromRequest(r)
	if err != nil {
		jsonOK(w, resp)
		return
	}
	info := &authSessionInfo{
		Sub:    claims.Sub,
		Name:   claims.Name,
		Method: claims.Method,
		Avatar: claims.Avatar,
		Owner:  claims.Owner,
	}
	if claims.Owner {
		for cap := range allCapabilities {
			info.Capabilities = append(info.Capabilities, string(cap))
		}
	} else {
		for cap := range capsForSession(claims) {
			info.Capabilities = append(info.Capabilities, string(cap))
		}
	}
	resp.Session = info
	jsonOK(w, resp)
}

// discordLoginConfigured reports whether Discord OAuth login is usable:
// explicitly enabled with OAuth app credentials plus the bot token + guild
// needed for the membership/role lookup.
func discordLoginConfigured(cfg appConfig) bool {
	return cfg.AuthDiscordEnabled != nil && *cfg.AuthDiscordEnabled &&
		cfg.AuthDiscordClientID != "" && cfg.AuthDiscordClientSecret != "" &&
		cfg.DiscordBotToken != "" && cfg.DiscordGuildID != ""
}

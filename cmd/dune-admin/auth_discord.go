package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	oauthStateCookieName = "dune_admin_oauth_state"
	// rolesMaxAge is how old a role snapshot may be before the middleware
	// lazily re-fetches it from Discord.
	rolesMaxAge = 15 * time.Minute
)

// errNotGuildMember marks a definitive "not in the guild" answer from
// Discord (HTTP 404), as opposed to transient API failures.
var errNotGuildMember = errors.New("not a member of the configured guild")

// discordMemberFetcher wraps the bot-token REST calls used to resolve a
// user's guild membership, roles, and the guild owner. Interface so tests
// inject a stub.
type discordMemberFetcher interface {
	fetchMember(guildID, userID string) (roleIDs []string, displayName string, err error)
	fetchGuildOwnerID(guildID string) (string, error)
}

// discordOAuthExchanger wraps the OAuth2 code exchange + identity fetch.
type discordOAuthExchanger interface {
	exchangeCode(code, redirectURI string) (accessToken string, err error)
	fetchSelf(accessToken string) (userID, username, avatarURL string, err error)
}

// discordAuthState holds the live Discord clients. Guarded because live
// config saves (request goroutines) replace them while other requests read.
var discordAuthState struct {
	mu        sync.RWMutex
	fetcher   discordMemberFetcher
	exchanger discordOAuthExchanger
}

func currentDiscordAuth() (discordMemberFetcher, discordOAuthExchanger) {
	discordAuthState.mu.RLock()
	defer discordAuthState.mu.RUnlock()
	return discordAuthState.fetcher, discordAuthState.exchanger
}

func setDiscordAuth(f discordMemberFetcher, e discordOAuthExchanger) {
	discordAuthState.mu.Lock()
	defer discordAuthState.mu.Unlock()
	discordAuthState.fetcher = f
	discordAuthState.exchanger = e
}

// initDiscordAuth wires (or tears down) the Discord clients for the current
// config. Called from initAuthRuntime at startup and after live config saves,
// so enabling/disabling Discord login in the UI takes effect immediately.
func initDiscordAuth(cfg appConfig) {
	if !discordLoginConfigured(cfg) {
		setDiscordAuth(nil, nil)
		return
	}
	sess, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		logAuthError("discord auth: bot session: " + err.Error())
		setDiscordAuth(nil, nil)
		return
	}
	// REST-only: never sess.Open() — no gateway connection is needed to
	// fetch members/roles, so login works even while the event bot is down.
	setDiscordAuth(
		&discordgoMemberFetcher{sess: sess},
		&discordHTTPExchanger{
			clientID:     cfg.AuthDiscordClientID,
			clientSecret: cfg.AuthDiscordClientSecret,
		},
	)
}

// ── real implementations ──────────────────────────────────────────────────

type discordgoMemberFetcher struct {
	sess *discordgo.Session
}

func (f *discordgoMemberFetcher) fetchMember(guildID, userID string) ([]string, string, error) {
	m, err := f.sess.GuildMember(guildID, userID)
	if err != nil {
		var restErr *discordgo.RESTError
		if errors.As(err, &restErr) && restErr.Response != nil && restErr.Response.StatusCode == http.StatusNotFound {
			return nil, "", errNotGuildMember
		}
		return nil, "", fmt.Errorf("fetch guild member: %w", err)
	}
	name := m.Nick
	if name == "" && m.User != nil {
		name = m.User.Username
	}
	return m.Roles, name, nil
}

func (f *discordgoMemberFetcher) fetchGuildOwnerID(guildID string) (string, error) {
	g, err := f.sess.Guild(guildID)
	if err != nil {
		return "", fmt.Errorf("fetch guild: %w", err)
	}
	return g.OwnerID, nil
}

type discordHTTPExchanger struct {
	clientID     string
	clientSecret string
}

func (e *discordHTTPExchanger) exchangeCode(code, redirectURI string) (string, error) {
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	}
	req, err := http.NewRequest(http.MethodPost, "https://discord.com/api/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(e.clientID, e.clientSecret)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange: discord returned %s", resp.Status)
	}
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if body.AccessToken == "" {
		return "", errors.New("token exchange: empty access token")
	}
	return body.AccessToken, nil
}

func (e *discordHTTPExchanger) fetchSelf(accessToken string) (string, string, string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://discord.com/api/users/@me", nil)
	if err != nil {
		return "", "", "", fmt.Errorf("build self request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("fetch self: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("fetch self: discord returned %s", resp.Status)
	}
	var body struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", "", fmt.Errorf("decode self response: %w", err)
	}
	avatarURL := ""
	if body.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png?size=64", body.ID, body.Avatar)
	}
	return body.ID, body.Username, avatarURL, nil
}

// ── owner resolution ──────────────────────────────────────────────────────

// resolveOwner reports whether a Discord user gets full (owner) rights:
// the guild owner, anyone in auth_owner_discord_ids, or anyone holding a
// role in auth_owner_role_ids.
func resolveOwner(userID string, roleIDs []string, guildOwnerID string, cfg appConfig) bool {
	if userID != "" && userID == guildOwnerID {
		return true
	}
	if slices.Contains(splitCommaList(cfg.AuthOwnerDiscordIDs), userID) {
		return true
	}
	ownerRoles := splitCommaList(cfg.AuthOwnerRoleIDs)
	return slices.ContainsFunc(roleIDs, func(role string) bool {
		return slices.Contains(ownerRoles, role)
	})
}

func splitCommaList(s string) []string {
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

// ── handlers ──────────────────────────────────────────────────────────────

// discordRedirectURI returns the OAuth redirect: the configured override or
// one derived from the request's scheme + host.
func discordRedirectURI(r *http.Request, cfg appConfig) string {
	if cfg.AuthDiscordRedirectURL != "" {
		return cfg.AuthDiscordRedirectURL
	}
	scheme := "http"
	if requestIsTLS(r) {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/api/v1/auth/discord/callback"
}

// handleAuthDiscordLogin starts the OAuth2 flow: sets a short-lived state
// cookie and redirects to Discord's authorize page.
//
// @Summary Start Discord OAuth2 login
// @Tags auth
// @Success 302
// @Failure 404 {object} map[string]string
// @Router /api/v1/auth/discord/login [get]
func handleAuthDiscordLogin(w http.ResponseWriter, r *http.Request) {
	cfg := loadedConfig
	if !authEnabled(cfg) || !discordLoginConfigured(cfg) {
		jsonErr(w, errors.New("discord login is not configured"), http.StatusNotFound)
		return
	}
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		jsonErr(w, errors.New("could not generate state"), http.StatusInternalServerError)
		return
	}
	state := hex.EncodeToString(stateBytes)
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/api/v1/auth/discord",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   requestIsTLS(r),
		SameSite: http.SameSiteLaxMode,
	})
	authorize := url.URL{
		Scheme: "https",
		Host:   "discord.com",
		Path:   "/oauth2/authorize",
		RawQuery: url.Values{
			"client_id":     {cfg.AuthDiscordClientID},
			"redirect_uri":  {discordRedirectURI(r, cfg)},
			"response_type": {"code"},
			"scope":         {"identify"},
			"state":         {state},
		}.Encode(),
	}
	http.Redirect(w, r, authorize.String(), http.StatusFound)
}

// handleAuthDiscordCallback completes the OAuth2 flow: validates state,
// exchanges the code, resolves guild membership + roles via the bot token,
// and mints the session cookie.
//
// @Summary Discord OAuth2 callback
// @Tags auth
// @Success 302
// @Failure 400 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Router /api/v1/auth/discord/callback [get]
func handleAuthDiscordCallback(w http.ResponseWriter, r *http.Request) {
	cfg := loadedConfig
	if !authEnabled(cfg) || !discordLoginConfigured(cfg) {
		jsonErr(w, errors.New("discord login is not configured"), http.StatusNotFound)
		return
	}
	fetcher, exchanger := currentDiscordAuth()
	if fetcher == nil || exchanger == nil {
		jsonErr(w, errors.New("discord auth not initialized"), http.StatusServiceUnavailable)
		return
	}

	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		jsonErr(w, errors.New("invalid OAuth state"), http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		jsonErr(w, errors.New("missing OAuth code"), http.StatusBadRequest)
		return
	}

	accessToken, err := exchanger.exchangeCode(code, discordRedirectURI(r, cfg))
	if err != nil {
		logAuthError("discord code exchange: " + err.Error())
		jsonErr(w, errors.New("discord token exchange failed"), http.StatusBadGateway)
		return
	}
	userID, username, avatarURL, err := exchanger.fetchSelf(accessToken)
	if err != nil {
		logAuthError("discord identity fetch: " + err.Error())
		jsonErr(w, errors.New("could not fetch discord identity"), http.StatusBadGateway)
		return
	}

	roleIDs, displayName, err := fetcher.fetchMember(cfg.DiscordGuildID, userID)
	if errors.Is(err, errNotGuildMember) {
		http.Redirect(w, r, "/#login-error=not-a-member", http.StatusFound)
		return
	}
	if err != nil {
		logAuthError("discord member fetch: " + err.Error())
		jsonErr(w, errors.New("could not verify guild membership"), http.StatusBadGateway)
		return
	}
	if displayName == "" {
		displayName = username
	}

	owner := resolveDiscordOwner(userID, roleIDs, cfg)

	secret := currentSessionSecret()
	if secret == nil {
		jsonErr(w, errors.New("session secret not initialized"), http.StatusInternalServerError)
		return
	}
	ttl := sessionTTL(cfg)
	token, err := mintSession(sessionClaims{
		Sub:     "discord:" + userID,
		Name:    displayName,
		Method:  "discord",
		Avatar:  avatarURL,
		RoleIDs: roleIDs,
		Owner:   owner,
		RolesAt: time.Now().Unix(),
	}, secret, ttl)
	if err != nil {
		jsonErr(w, errors.New("could not create session"), http.StatusInternalServerError)
		return
	}
	setSessionCookie(w, r, token, ttl)
	http.Redirect(w, r, "/", http.StatusFound)
}

// resolveDiscordOwner combines the guild-owner lookup (best effort) with the
// configured owner lists. A failed guild fetch only disables the implicit
// guild-owner grant — configured lists still apply.
func resolveDiscordOwner(userID string, roleIDs []string, cfg appConfig) bool {
	guildOwnerID := ""
	if fetcher, _ := currentDiscordAuth(); fetcher != nil {
		id, err := fetcher.fetchGuildOwnerID(cfg.DiscordGuildID)
		if err != nil {
			logAuthError("guild owner lookup: " + err.Error())
		} else {
			guildOwnerID = id
		}
	}
	return resolveOwner(userID, roleIDs, guildOwnerID, cfg)
}

// refreshDiscordSession lazily re-fetches the role snapshot for Discord
// sessions older than rolesMaxAge. Definitive non-membership invalidates the
// session; transient Discord failures keep the stale snapshot working.
func refreshDiscordSession(w http.ResponseWriter, r *http.Request, claims *sessionClaims) (*sessionClaims, error) {
	if claims.Method != "discord" {
		return claims, nil
	}
	if time.Since(time.Unix(claims.RolesAt, 0)) < rolesMaxAge {
		return claims, nil
	}
	cfg := loadedConfig
	fetcher, _ := currentDiscordAuth()
	if fetcher == nil {
		return claims, nil
	}
	userID := strings.TrimPrefix(claims.Sub, "discord:")
	roleIDs, displayName, err := fetcher.fetchMember(cfg.DiscordGuildID, userID)
	if errors.Is(err, errNotGuildMember) {
		return nil, errNotGuildMember
	}
	if err != nil {
		logAuthError("role refresh (keeping stale snapshot): " + err.Error())
		return claims, nil
	}
	if displayName == "" {
		displayName = claims.Name
	}

	updated := *claims
	updated.RoleIDs = roleIDs
	updated.Name = displayName
	updated.Owner = resolveDiscordOwner(userID, roleIDs, cfg)
	updated.RolesAt = time.Now().Unix()

	secret := currentSessionSecret()
	if secret == nil {
		return claims, nil
	}
	ttl := sessionTTL(cfg)
	token, err := mintSession(updated, secret, ttl)
	if err != nil {
		logAuthError("session re-mint: " + err.Error())
		return claims, nil
	}
	setSessionCookie(w, r, token, ttl)
	return &updated, nil
}

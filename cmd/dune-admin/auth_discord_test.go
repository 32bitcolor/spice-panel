package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// stubDiscordAuth implements discordMemberFetcher + discordOAuthExchanger.
type stubDiscordAuth struct {
	memberRoles  []string
	memberName   string
	memberErr    error
	guildOwnerID string
	ownerErr     error

	exchangeToken string
	exchangeErr   error
	selfID        string
	selfName      string
	selfAvatar    string
	selfErr       error
}

func (s *stubDiscordAuth) fetchMember(_, _ string) ([]string, string, error) {
	return s.memberRoles, s.memberName, s.memberErr
}

func (s *stubDiscordAuth) fetchGuildOwnerID(_ string) (string, error) {
	return s.guildOwnerID, s.ownerErr
}

func (s *stubDiscordAuth) exchangeCode(_, _ string) (string, error) {
	return s.exchangeToken, s.exchangeErr
}

func (s *stubDiscordAuth) fetchSelf(_ string) (string, string, string, error) {
	return s.selfID, s.selfName, s.selfAvatar, s.selfErr
}

func discordAuthConfig() appConfig {
	enabled := true
	return appConfig{
		AuthEnabled:             &enabled,
		AuthDiscordEnabled:      &enabled,
		AuthDiscordClientID:     "client123",
		AuthDiscordClientSecret: "secret456",
		DiscordBotToken:         "bot-token",
		DiscordGuildID:          "guild789",
	}
}

func withStubDiscord(t *testing.T, stub *stubDiscordAuth) {
	t.Helper()
	oldFetcher, oldExchanger := currentDiscordAuth()
	if stub != nil {
		setDiscordAuth(stub, stub)
	}
	t.Cleanup(func() {
		setDiscordAuth(oldFetcher, oldExchanger)
	})
}

func TestInitAuthRuntimeReappliesDiscordAuth(t *testing.T) {
	t.Setenv("DUNE_ADMIN_CONFIG_DIR", t.TempDir())
	enabled := true
	secret := []byte("0123456789abcdef0123456789abcdef")

	// Boot with local-only auth: no Discord client wired.
	withAuthTestConfig(t, appConfig{
		AuthEnabled: &enabled, AuthLocalUsername: "a", AuthLocalPasswordHash: "x",
	}, secret)
	withStubDiscord(t, nil) // isolate + restore package state
	setDiscordAuth(nil, nil)
	initAuthRuntime(loadedConfig)
	if f, _ := currentDiscordAuth(); f != nil {
		t.Fatal("discord auth wired without discord config")
	}

	// Live config save adds Discord login: must wire it up even though the
	// session secret is already initialized.
	loadedConfig = discordAuthConfig()
	initAuthRuntime(loadedConfig)
	if f, e := currentDiscordAuth(); f == nil || e == nil {
		t.Fatal("discord auth not initialized after live config change")
	}

	// Live config save removes Discord login: must tear it down.
	loadedConfig = appConfig{AuthEnabled: &enabled, AuthLocalUsername: "a", AuthLocalPasswordHash: "x"}
	initAuthRuntime(loadedConfig)
	if f, _ := currentDiscordAuth(); f != nil {
		t.Fatal("discord auth not torn down after disabling")
	}
}

func TestResolveOwner(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		roleIDs []string
		ownerID string
		cfg     appConfig
		want    bool
	}{
		{"guild owner", "111", nil, "111", appConfig{}, true},
		{"not owner", "222", nil, "111", appConfig{}, false},
		{"configured owner id", "333", nil, "111", appConfig{AuthOwnerDiscordIDs: "333,444"}, true},
		{"configured owner role", "555", []string{"role9"}, "111", appConfig{AuthOwnerRoleIDs: "role8, role9"}, true},
		{"role not in owner roles", "555", []string{"role7"}, "111", appConfig{AuthOwnerRoleIDs: "role8,role9"}, false},
		{"empty config lists", "666", []string{"r"}, "111", appConfig{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveOwner(tt.userID, tt.roleIDs, tt.ownerID, tt.cfg); got != tt.want {
				t.Errorf("resolveOwner = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleAuthDiscordLogin(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")

	t.Run("redirects to discord authorize with state cookie", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		r := httptest.NewRequest(http.MethodGet, "http://dash.example.com/api/v1/auth/discord/login", nil)
		w := httptest.NewRecorder()
		handleAuthDiscordLogin(w, r)
		if w.Code != http.StatusFound {
			t.Fatalf("status = %d, want 302 (body %s)", w.Code, w.Body.String())
		}
		loc, err := url.Parse(w.Header().Get("Location"))
		if err != nil {
			t.Fatal(err)
		}
		if loc.Host != "discord.com" {
			t.Errorf("redirect host = %q", loc.Host)
		}
		q := loc.Query()
		if q.Get("client_id") != "client123" {
			t.Errorf("client_id = %q", q.Get("client_id"))
		}
		if q.Get("scope") != "identify" {
			t.Errorf("scope = %q", q.Get("scope"))
		}
		if !strings.Contains(q.Get("redirect_uri"), "/api/v1/auth/discord/callback") {
			t.Errorf("redirect_uri = %q", q.Get("redirect_uri"))
		}
		state := q.Get("state")
		if state == "" {
			t.Fatal("no state in authorize URL")
		}
		var stateCookie *http.Cookie
		for _, c := range w.Result().Cookies() {
			if c.Name == oauthStateCookieName {
				stateCookie = c
			}
		}
		if stateCookie == nil {
			t.Fatal("state cookie not set")
		}
		if stateCookie.Value != state {
			t.Error("state cookie does not match state param")
		}
		if !stateCookie.HttpOnly {
			t.Error("state cookie not HttpOnly")
		}
	})

	t.Run("404 when discord login not configured", func(t *testing.T) {
		enabled := true
		withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, secret)
		r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/discord/login", nil)
		w := httptest.NewRecorder()
		handleAuthDiscordLogin(w, r)
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})
}

func callbackRequest(state, cookieState string) *http.Request {
	r := httptest.NewRequest(http.MethodGet,
		"http://dash.example.com/api/v1/auth/discord/callback?code=authcode&state="+state, nil)
	if cookieState != "" {
		r.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: cookieState})
	}
	return r
}

func TestHandleAuthDiscordCallback(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")

	happyStub := func() *stubDiscordAuth {
		return &stubDiscordAuth{
			exchangeToken: "user-token",
			selfID:        "42",
			selfName:      "PlayerOne",
			selfAvatar:    "https://cdn.discordapp.com/avatars/42/abc123.png",
			memberRoles:   []string{"role1", "role2"},
			memberName:    "NickName",
			guildOwnerID:  "999",
		}
	}

	t.Run("member gets session cookie with role snapshot", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, happyStub())
		w := httptest.NewRecorder()
		handleAuthDiscordCallback(w, callbackRequest("st", "st"))
		if w.Code != http.StatusFound {
			t.Fatalf("status = %d (body %s)", w.Code, w.Body.String())
		}
		if loc := w.Header().Get("Location"); loc != "/" {
			t.Errorf("redirect = %q, want /", loc)
		}
		var sess string
		for _, c := range w.Result().Cookies() {
			if c.Name == sessionCookieName {
				sess = c.Value
			}
		}
		if sess == "" {
			t.Fatal("no session cookie")
		}
		claims, err := verifySession(sess, secret)
		if err != nil {
			t.Fatal(err)
		}
		if claims.Sub != "discord:42" || claims.Method != "discord" {
			t.Errorf("claims = %+v", claims)
		}
		if len(claims.RoleIDs) != 2 {
			t.Errorf("roles = %v", claims.RoleIDs)
		}
		if claims.Owner {
			t.Error("non-owner marked owner")
		}
		if claims.RolesAt == 0 {
			t.Error("RolesAt not stamped")
		}
		if claims.Avatar != "https://cdn.discordapp.com/avatars/42/abc123.png" {
			t.Errorf("avatar = %q", claims.Avatar)
		}
	})

	t.Run("guild owner marked owner", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		stub := happyStub()
		stub.guildOwnerID = "42"
		withStubDiscord(t, stub)
		w := httptest.NewRecorder()
		handleAuthDiscordCallback(w, callbackRequest("st", "st"))
		for _, c := range w.Result().Cookies() {
			if c.Name == sessionCookieName {
				claims, err := verifySession(c.Value, secret)
				if err != nil {
					t.Fatal(err)
				}
				if !claims.Owner {
					t.Error("guild owner not marked owner")
				}
				return
			}
		}
		t.Fatal("no session cookie")
	})

	t.Run("non-member redirected with error", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		stub := happyStub()
		stub.memberErr = errNotGuildMember
		withStubDiscord(t, stub)
		w := httptest.NewRecorder()
		handleAuthDiscordCallback(w, callbackRequest("st", "st"))
		if w.Code != http.StatusFound {
			t.Fatalf("status = %d", w.Code)
		}
		if loc := w.Header().Get("Location"); !strings.Contains(loc, "login-error=not-a-member") {
			t.Errorf("redirect = %q, want not-a-member error", loc)
		}
		for _, c := range w.Result().Cookies() {
			if c.Name == sessionCookieName {
				t.Error("session cookie set for non-member")
			}
		}
	})

	t.Run("state mismatch → 400", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, happyStub())
		w := httptest.NewRecorder()
		handleAuthDiscordCallback(w, callbackRequest("st", "other"))
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("missing state cookie → 400", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, happyStub())
		w := httptest.NewRecorder()
		handleAuthDiscordCallback(w, callbackRequest("st", ""))
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("exchange failure → 502", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		stub := happyStub()
		stub.exchangeErr = errors.New("discord down")
		withStubDiscord(t, stub)
		w := httptest.NewRecorder()
		handleAuthDiscordCallback(w, callbackRequest("st", "st"))
		if w.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502", w.Code)
		}
	})
}

func TestRefreshDiscordSession(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	freshClaims := func(age time.Duration) *sessionClaims {
		return &sessionClaims{
			Sub: "discord:42", Name: "PlayerOne", Method: "discord",
			RoleIDs: []string{"old-role"},
			RolesAt: time.Now().Add(-age).Unix(),
		}
	}

	t.Run("local sessions never refreshed", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, &stubDiscordAuth{memberErr: errors.New("must not be called")})
		claims := &sessionClaims{Sub: "local:admin", Method: "local", Owner: true}
		got, err := refreshDiscordSession(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), claims)
		if err != nil || got != claims {
			t.Errorf("local session altered: %v %v", got, err)
		}
	})

	t.Run("fresh snapshot untouched", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, &stubDiscordAuth{memberErr: errors.New("must not be called")})
		claims := freshClaims(time.Minute)
		got, err := refreshDiscordSession(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), claims)
		if err != nil || got != claims {
			t.Errorf("fresh session refreshed: %v %v", got, err)
		}
	})

	t.Run("stale snapshot refreshed and cookie re-minted", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, &stubDiscordAuth{
			memberRoles: []string{"new-role"}, memberName: "PlayerOne", guildOwnerID: "999",
		})
		w := httptest.NewRecorder()
		got, err := refreshDiscordSession(w, httptest.NewRequest("GET", "/", nil), freshClaims(time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if len(got.RoleIDs) != 1 || got.RoleIDs[0] != "new-role" {
			t.Errorf("roles = %v, want [new-role]", got.RoleIDs)
		}
		found := false
		for _, c := range w.Result().Cookies() {
			if c.Name == sessionCookieName {
				found = true
			}
		}
		if !found {
			t.Error("refreshed session cookie not re-set")
		}
	})

	t.Run("kicked member → error", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, &stubDiscordAuth{memberErr: errNotGuildMember})
		_, err := refreshDiscordSession(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), freshClaims(time.Hour))
		if err == nil {
			t.Error("kicked member session survived")
		}
	})

	t.Run("transient discord error keeps stale roles", func(t *testing.T) {
		withAuthTestConfig(t, discordAuthConfig(), secret)
		withStubDiscord(t, &stubDiscordAuth{memberErr: errors.New("503 from discord")})
		claims := freshClaims(time.Hour)
		got, err := refreshDiscordSession(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), claims)
		if err != nil {
			t.Fatalf("transient error killed session: %v", err)
		}
		if len(got.RoleIDs) != 1 || got.RoleIDs[0] != "old-role" {
			t.Errorf("stale roles not preserved: %v", got.RoleIDs)
		}
	})
}

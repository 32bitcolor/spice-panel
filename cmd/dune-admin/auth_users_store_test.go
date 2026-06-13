package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestAuthUserStore(t *testing.T) *authUserStore {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := initAuthUsersSchema(db); err != nil {
		t.Fatal(err)
	}
	return newAuthUserStore(db)
}

func withTestAuthUsers(t *testing.T, s *authUserStore) {
	t.Helper()
	old := authUsersDB
	authUsersDB = s
	t.Cleanup(func() { authUsersDB = old })
}

func TestAuthUserStore(t *testing.T) {
	t.Run("upsert, list, verify round trip", func(t *testing.T) {
		s := newTestAuthUserStore(t)
		hash, _ := hashPassword("pw1")
		if err := s.upsert("mod", hash, []string{"players:write", "logs:read"}, true); err != nil {
			t.Fatal(err)
		}
		users, err := s.list()
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 1 || users[0].Username != "mod" || !users[0].Enabled {
			t.Fatalf("users = %+v", users)
		}
		if len(users[0].Capabilities) != 2 {
			t.Errorf("caps = %v", users[0].Capabilities)
		}
		caps, ok := s.verify("mod", "pw1")
		if !ok {
			t.Fatal("valid credentials rejected")
		}
		if len(caps) != 2 {
			t.Errorf("verify caps = %v", caps)
		}
		if _, ok := s.verify("mod", "wrong"); ok {
			t.Error("wrong password accepted")
		}
		if _, ok := s.verify("ghost", "pw1"); ok {
			t.Error("unknown user accepted")
		}
	})

	t.Run("update without password keeps old hash", func(t *testing.T) {
		s := newTestAuthUserStore(t)
		hash, _ := hashPassword("pw1")
		_ = s.upsert("mod", hash, []string{"players:write"}, true)
		if err := s.upsert("mod", "", []string{"logs:read"}, true); err != nil {
			t.Fatal(err)
		}
		if _, ok := s.verify("mod", "pw1"); !ok {
			t.Error("password lost on capability-only update")
		}
		caps, _ := s.verify("mod", "pw1")
		if len(caps) != 1 || caps[0] != "logs:read" {
			t.Errorf("caps not updated: %v", caps)
		}
	})

	t.Run("disabled user cannot verify", func(t *testing.T) {
		s := newTestAuthUserStore(t)
		hash, _ := hashPassword("pw1")
		_ = s.upsert("mod", hash, nil, false)
		if _, ok := s.verify("mod", "pw1"); ok {
			t.Error("disabled user verified")
		}
	})

	t.Run("delete removes user", func(t *testing.T) {
		s := newTestAuthUserStore(t)
		hash, _ := hashPassword("pw1")
		_ = s.upsert("mod", hash, nil, true)
		if err := s.deleteUser("mod"); err != nil {
			t.Fatal(err)
		}
		users, _ := s.list()
		if len(users) != 0 {
			t.Errorf("users = %+v", users)
		}
	})

	t.Run("new user without password rejected", func(t *testing.T) {
		s := newTestAuthUserStore(t)
		if err := s.upsert("mod", "", nil, true); err == nil {
			t.Error("user created without password hash")
		}
	})
}

func TestDBUserLogin(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	s := newTestAuthUserStore(t)
	hash, _ := hashPassword("modpw")
	_ = s.upsert("mod", hash, []string{"players:write"}, true)
	withTestAuthUsers(t, s)
	withAuthTestConfig(t, localAuthConfig(t, "adminpw"), secret)
	resetLoginLimiter()

	t.Run("db user can log in, is not owner", func(t *testing.T) {
		w := postLogin(t, `{"username":"mod","password":"modpw"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		claims, err := verifySession(w.Result().Cookies()[0].Value, secret)
		if err != nil {
			t.Fatal(err)
		}
		if claims.Owner {
			t.Error("db user must not be owner")
		}
		if claims.Sub != "local:mod" {
			t.Errorf("sub = %q", claims.Sub)
		}
	})

	t.Run("db user gets exactly their assigned caps plus default cascade", func(t *testing.T) {
		withTestMatrix(t, map[string][]string{"default": {"players:read"}})
		caps := capsForSession(&sessionClaims{Sub: "local:mod", Method: "local"})
		if !caps[capPlayersWrite] {
			t.Error("assigned capability missing")
		}
		if !caps[capPlayersRead] {
			t.Error("default cascade capability missing")
		}
		if caps[capLogsRead] {
			t.Error("unassigned capability granted")
		}
	})

	t.Run("config admin still works and is owner", func(t *testing.T) {
		resetLoginLimiter()
		w := postLogin(t, `{"username":"admin","password":"adminpw"}`)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		claims, _ := verifySession(w.Result().Cookies()[0].Value, secret)
		if !claims.Owner {
			t.Error("config admin must be owner")
		}
	})
}

func TestAuthUsersHandlers(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	enabled := true
	cfg := appConfig{AuthEnabled: &enabled}

	adminReq := func(t *testing.T, method, path, body string, owner bool, caps []string) *http.Request {
		t.Helper()
		claims := sessionClaims{Sub: "local:x", Method: "discord", Owner: owner}
		if len(caps) > 0 {
			claims.Method = "discord"
			claims.RoleIDs = []string{"admin-role"}
		}
		tok := mintTestSession(t, claims, secret)
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tok})
		if method != http.MethodGet {
			r.SetPathValue("username", strings.TrimPrefix(path, "/api/v1/auth/users/"))
		}
		return r
	}

	t.Run("owner can create and list users", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		withTestAuthUsers(t, newTestAuthUserStore(t))
		w := httptest.NewRecorder()
		handlePutAuthUser(w, adminReq(t, "PUT", "/api/v1/auth/users/mod",
			`{"password":"modpw","capabilities":["players:write"],"enabled":true}`, true, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("create status = %d (%s)", w.Code, w.Body.String())
		}
		w = httptest.NewRecorder()
		handleListAuthUsers(w, adminReq(t, "GET", "/api/v1/auth/users", "", true, nil))
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "mod") {
			t.Errorf("list = %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("auth:manage role can manage users", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		withTestAuthUsers(t, newTestAuthUserStore(t))
		withTestMatrix(t, map[string][]string{"admin-role": {"auth:manage"}})
		w := httptest.NewRecorder()
		handlePutAuthUser(w, adminReq(t, "PUT", "/api/v1/auth/users/mod",
			`{"password":"modpw","capabilities":[],"enabled":true}`, false, []string{"auth:manage"}))
		if w.Code != http.StatusOK {
			t.Errorf("status = %d (%s)", w.Code, w.Body.String())
		}
	})

	t.Run("non-admin denied", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		withTestAuthUsers(t, newTestAuthUserStore(t))
		withTestMatrix(t, map[string][]string{})
		w := httptest.NewRecorder()
		handleListAuthUsers(w, adminReq(t, "GET", "/api/v1/auth/users", "", false, nil))
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", w.Code)
		}
	})

	t.Run("invalid capability rejected", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		withTestAuthUsers(t, newTestAuthUserStore(t))
		w := httptest.NewRecorder()
		handlePutAuthUser(w, adminReq(t, "PUT", "/api/v1/auth/users/mod",
			`{"password":"pw","capabilities":["bogus:cap"],"enabled":true}`, true, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("delete user", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		s := newTestAuthUserStore(t)
		hash, _ := hashPassword("pw")
		_ = s.upsert("mod", hash, nil, true)
		withTestAuthUsers(t, s)
		w := httptest.NewRecorder()
		handleDeleteAuthUser(w, adminReq(t, "DELETE", "/api/v1/auth/users/mod", "", true, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		users, _ := s.list()
		if len(users) != 0 {
			t.Error("user not deleted")
		}
	})
}

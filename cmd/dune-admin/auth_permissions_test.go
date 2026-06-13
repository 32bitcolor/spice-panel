package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func withTestMatrix(t *testing.T, m map[string][]string) {
	t.Helper()
	old := snapshotPermissionsMatrix()
	setPermissionsMatrix(m)
	t.Cleanup(func() { setPermissionsMatrix(old) })
}

func TestCapsForRoles(t *testing.T) {
	withTestMatrix(t, map[string][]string{
		"mod-role":   {"players:write", "logs:read"},
		"admin-role": {"server:control", "config:write"},
	})

	t.Run("mapped role grants exactly its caps, nothing implicit", func(t *testing.T) {
		caps := capsForRoles([]string{"mod-role"})
		if !caps[capPlayersWrite] || !caps[capLogsRead] {
			t.Errorf("mapped caps missing: %v", caps)
		}
		if caps[capPlayersRead] {
			t.Error("players:read granted implicitly — no floor must exist")
		}
		if caps[capServerControl] {
			t.Error("got capability from unheld role")
		}
	})

	t.Run("multiple roles union", func(t *testing.T) {
		caps := capsForRoles([]string{"mod-role", "admin-role"})
		if !caps[capPlayersWrite] || !caps[capServerControl] || !caps[capConfigWrite] {
			t.Errorf("union incomplete: %v", caps)
		}
	})

	t.Run("unmapped roles get nothing", func(t *testing.T) {
		caps := capsForRoles([]string{"random-role"})
		if len(caps) != 0 {
			t.Errorf("caps = %v, want empty (deny by default)", caps)
		}
	})

	t.Run("default pseudo-row cascades to every member", func(t *testing.T) {
		withTestMatrix(t, map[string][]string{"default": {"events:read"}})
		caps := capsForRoles([]string{"any-role"})
		if !caps[capEventsRead] {
			t.Error("default pseudo-row not applied")
		}
	})

	t.Run("default cascades to guests too; guest row adds more", func(t *testing.T) {
		withTestMatrix(t, map[string][]string{"guest": {"events:read"}, "default": {"logs:read"}})
		guestCaps := capsForSession(&sessionClaims{Method: "guest"})
		if !guestCaps[capEventsRead] {
			t.Error("guest pseudo-row not applied to guest")
		}
		if !guestCaps[capLogsRead] {
			t.Error("default pseudo-row must cascade to guests")
		}
		memberCaps := capsForSession(&sessionClaims{Method: "discord"})
		if memberCaps[capEventsRead] {
			t.Error("guest pseudo-row must not apply to members")
		}
		if !memberCaps[capLogsRead] {
			t.Error("default pseudo-row must cascade to members")
		}
	})

	t.Run("unknown capability strings ignored", func(t *testing.T) {
		withTestMatrix(t, map[string][]string{"r": {"bogus:cap", "players:write"}})
		caps := capsForRoles([]string{"r"})
		if !caps[capPlayersWrite] {
			t.Error("valid capability dropped")
		}
		if caps[capability("bogus:cap")] {
			t.Error("bogus capability honored")
		}
	})
}

func TestPermissionsMatrixPersistence(t *testing.T) {
	t.Run("save and load round trip", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "permissions.yaml")
		m := map[string][]string{"role1": {"players:write"}, "role2": {"logs:read", "data:export"}}
		if err := savePermissionsMatrix(path, m); err != nil {
			t.Fatalf("save: %v", err)
		}
		got, err := loadPermissionsMatrix(path)
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if len(got) != 2 || len(got["role2"]) != 2 {
			t.Errorf("got %v", got)
		}
		info, _ := os.Stat(path)
		if info.Mode().Perm() != 0o600 {
			t.Errorf("mode = %v, want 0600", info.Mode().Perm())
		}
	})

	t.Run("missing file returns empty matrix", func(t *testing.T) {
		got, err := loadPermissionsMatrix(filepath.Join(t.TempDir(), "nope.yaml"))
		if err != nil {
			t.Fatalf("missing file should not error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})

	t.Run("corrupt file errors", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "permissions.yaml")
		if err := os.WriteFile(path, []byte("{{{not yaml"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadPermissionsMatrix(path); err == nil {
			t.Error("corrupt file loaded without error")
		}
	})
}

func TestConcurrentMatrixAccess(t *testing.T) {
	withTestMatrix(t, map[string][]string{"r": {"players:write"}})
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			capsForRoles([]string{"r"})
		}()
		go func() {
			defer wg.Done()
			setPermissionsMatrix(map[string][]string{"r": {"logs:read"}})
		}()
	}
	wg.Wait()
}

func ownerRequest(t *testing.T, method, path, body string, secret []byte, owner bool) *http.Request {
	t.Helper()
	claims := sessionClaims{Sub: "discord:1", Method: "discord", Owner: owner}
	tok := mintTestSession(t, claims, secret)
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tok})
	return r
}

func TestPermissionsHandlers(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	enabled := true
	cfg := appConfig{AuthEnabled: &enabled}

	t.Run("GET requires owner", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		w := httptest.NewRecorder()
		handleGetPermissions(w, ownerRequest(t, "GET", "/api/v1/auth/permissions", "", secret, false))
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403 for non-owner", w.Code)
		}
	})

	t.Run("GET returns matrix and capability list", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		withTestMatrix(t, map[string][]string{"role1": {"players:write"}})
		w := httptest.NewRecorder()
		handleGetPermissions(w, ownerRequest(t, "GET", "/api/v1/auth/permissions", "", secret, true))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		var resp struct {
			Matrix       map[string][]string `json:"matrix"`
			Capabilities []struct {
				ID          string `json:"id"`
				Description string `json:"description"`
			} `json:"capabilities"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Matrix["role1"]) != 1 {
			t.Errorf("matrix = %v", resp.Matrix)
		}
		if len(resp.Capabilities) != len(allCapabilities) {
			t.Errorf("capabilities = %d, want %d", len(resp.Capabilities), len(allCapabilities))
		}
	})

	t.Run("PUT validates capabilities", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		tmp := filepath.Join(t.TempDir(), "permissions.yaml")
		withPermissionsPath(t, tmp)
		w := httptest.NewRecorder()
		handlePutPermissions(w, ownerRequest(t, "PUT", "/api/v1/auth/permissions",
			`{"matrix":{"role1":["bogus:cap"]}}`, secret, true))
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400 for unknown capability", w.Code)
		}
	})

	t.Run("PUT persists and applies live", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		withTestMatrix(t, map[string][]string{})
		tmp := filepath.Join(t.TempDir(), "permissions.yaml")
		withPermissionsPath(t, tmp)
		w := httptest.NewRecorder()
		handlePutPermissions(w, ownerRequest(t, "PUT", "/api/v1/auth/permissions",
			`{"matrix":{"role1":["players:write","logs:read"]}}`, secret, true))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		// Applied live:
		caps := capsForRoles([]string{"role1"})
		if !caps[capPlayersWrite] || !caps[capLogsRead] {
			t.Errorf("live matrix not applied: %v", caps)
		}
		// Persisted:
		loaded, err := loadPermissionsMatrix(tmp)
		if err != nil || len(loaded["role1"]) != 2 {
			t.Errorf("persisted matrix = %v (err %v)", loaded, err)
		}
	})

	t.Run("PUT requires owner", func(t *testing.T) {
		withAuthTestConfig(t, cfg, secret)
		w := httptest.NewRecorder()
		handlePutPermissions(w, ownerRequest(t, "PUT", "/api/v1/auth/permissions",
			`{"matrix":{}}`, secret, false))
		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want 403", w.Code)
		}
	})
}

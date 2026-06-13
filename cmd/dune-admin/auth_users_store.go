package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// authUserStore persists local dashboard accounts (beyond the config.yaml
// bootstrap admin) in the unified dune-admin.db SQLite store. Each user has
// a bcrypt hash and a directly-assigned capability list.
type authUserStore struct {
	db *sql.DB
}

const authUsersSchema = `
CREATE TABLE IF NOT EXISTS auth_users (
	username      TEXT PRIMARY KEY,
	password_hash TEXT NOT NULL,
	capabilities  TEXT NOT NULL DEFAULT '[]',
	enabled       INTEGER NOT NULL DEFAULT 1,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL
);`

// initAuthUsersSchema creates the auth_users table. Idempotent.
func initAuthUsersSchema(db *sql.DB) error {
	if _, err := db.Exec(authUsersSchema); err != nil {
		return fmt.Errorf("init auth users schema: %w", err)
	}
	return nil
}

func newAuthUserStore(db *sql.DB) *authUserStore {
	return &authUserStore{db: db}
}

// authUsersDB is the live store handle, wired at startup from globalStore.
// Nil when the unified store is unavailable — DB-user login is then off.
var authUsersDB *authUserStore

type authUserRecord struct {
	Username     string   `json:"username"`
	Capabilities []string `json:"capabilities"`
	Enabled      bool     `json:"enabled"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// list returns all local users without their password hashes.
func (s *authUserStore) list() ([]authUserRecord, error) {
	rows, err := s.db.Query(`SELECT username, capabilities, enabled, created_at, updated_at FROM auth_users ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("list auth users: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []authUserRecord
	for rows.Next() {
		var rec authUserRecord
		var capsJSON string
		var enabled int
		if err := rows.Scan(&rec.Username, &capsJSON, &enabled, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan auth user: %w", err)
		}
		rec.Enabled = enabled != 0
		if err := json.Unmarshal([]byte(capsJSON), &rec.Capabilities); err != nil {
			rec.Capabilities = nil
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// upsert creates or updates a user. An empty hash on an existing user keeps
// the current password; on a new user it is an error.
func (s *authUserStore) upsert(username, hash string, caps []string, enabled bool) error {
	if username == "" {
		return errors.New("username must not be empty")
	}
	if caps == nil {
		caps = []string{}
	}
	capsJSON, err := json.Marshal(caps)
	if err != nil {
		return fmt.Errorf("marshal capabilities: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	if hash == "" {
		res, err := s.db.Exec(
			`UPDATE auth_users SET capabilities = ?, enabled = ?, updated_at = ? WHERE username = ?`,
			string(capsJSON), enabledInt, now, username)
		if err != nil {
			return fmt.Errorf("update auth user: %w", err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return errors.New("new users require a password")
		}
		return nil
	}
	_, err = s.db.Exec(
		`INSERT INTO auth_users (username, password_hash, capabilities, enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(username) DO UPDATE SET
		   password_hash = excluded.password_hash,
		   capabilities  = excluded.capabilities,
		   enabled       = excluded.enabled,
		   updated_at    = excluded.updated_at`,
		username, hash, string(capsJSON), enabledInt, now, now)
	if err != nil {
		return fmt.Errorf("upsert auth user: %w", err)
	}
	return nil
}

func (s *authUserStore) deleteUser(username string) error {
	if _, err := s.db.Exec(`DELETE FROM auth_users WHERE username = ?`, username); err != nil {
		return fmt.Errorf("delete auth user: %w", err)
	}
	return nil
}

// verify checks credentials for an enabled user, returning their assigned
// capabilities on success.
func (s *authUserStore) verify(username, password string) ([]string, bool) {
	var hash, capsJSON string
	var enabled int
	err := s.db.QueryRow(
		`SELECT password_hash, capabilities, enabled FROM auth_users WHERE username = ?`,
		username).Scan(&hash, &capsJSON, &enabled)
	if err != nil || enabled == 0 {
		return nil, false
	}
	if !checkPassword(hash, password) {
		return nil, false
	}
	var caps []string
	_ = json.Unmarshal([]byte(capsJSON), &caps)
	return caps, true
}

// capsForUser returns the stored capability list for an enabled user.
func (s *authUserStore) capsForUser(username string) []string {
	var capsJSON string
	var enabled int
	err := s.db.QueryRow(
		`SELECT capabilities, enabled FROM auth_users WHERE username = ?`,
		username).Scan(&capsJSON, &enabled)
	if err != nil || enabled == 0 {
		return nil
	}
	var caps []string
	_ = json.Unmarshal([]byte(capsJSON), &caps)
	return caps
}

// ── owner / auth:manage gated endpoints ───────────────────────────────────

// requireAuthAdmin allows owners and sessions holding auth:manage.
func requireAuthAdmin(w http.ResponseWriter, r *http.Request) (*sessionClaims, bool) {
	claims, err := sessionFromRequest(r)
	if err != nil {
		jsonErr(w, errors.New("authentication required"), http.StatusUnauthorized)
		return nil, false
	}
	if !claims.Owner && !capsForSession(claims)[capAuthManage] {
		jsonErr(w, errors.New("permission management access required"), http.StatusForbidden)
		return nil, false
	}
	return claims, true
}

// handleListAuthUsers returns all local dashboard users.
//
// @Summary List local dashboard users (owner / auth:manage)
// @Tags auth
// @Produce json
// @Success 200 {array} authUserRecord
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/auth/users [get]
func handleListAuthUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthAdmin(w, r); !ok {
		return
	}
	if authUsersDB == nil {
		jsonErr(w, errors.New("local user store unavailable"), http.StatusServiceUnavailable)
		return
	}
	users, err := authUsersDB.list()
	if err != nil {
		jsonErr(w, errors.New("could not list users"), http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []authUserRecord{}
	}
	jsonOK(w, users)
}

type authUserUpdateRequest struct {
	Password     string   `json:"password,omitempty"`
	Capabilities []string `json:"capabilities"`
	Enabled      bool     `json:"enabled"`
}

// handlePutAuthUser creates or updates a local dashboard user.
//
// @Summary Create or update a local dashboard user (owner / auth:manage)
// @Tags auth
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param user body authUserUpdateRequest true "User settings"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/auth/users/{username} [put]
func handlePutAuthUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireAuthAdmin(w, r)
	if !ok {
		return
	}
	if authUsersDB == nil {
		jsonErr(w, errors.New("local user store unavailable"), http.StatusServiceUnavailable)
		return
	}
	username := r.PathValue("username")
	var req authUserUpdateRequest
	if err := decode(r, &req); err != nil {
		jsonErr(w, fmt.Errorf("decode: %w", err), http.StatusBadRequest)
		return
	}
	for _, name := range req.Capabilities {
		if _, ok := allCapabilities[capability(name)]; !ok {
			jsonErr(w, fmt.Errorf("unknown capability %q", name), http.StatusBadRequest)
			return
		}
	}
	hash := ""
	if req.Password != "" {
		h, err := hashPassword(req.Password)
		if err != nil {
			jsonErr(w, errors.New("could not hash password"), http.StatusInternalServerError)
			return
		}
		hash = h
	}
	if err := authUsersDB.upsert(username, hash, req.Capabilities, req.Enabled); err != nil {
		jsonErr(w, err, http.StatusBadRequest)
		return
	}
	if sink := currentAuditSink(); sink != nil {
		sink(claims, r, http.StatusOK)
	}
	jsonOK(w, map[string]string{"status": "saved"})
}

// handleDeleteAuthUser removes a local dashboard user.
//
// @Summary Delete a local dashboard user (owner / auth:manage)
// @Tags auth
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/auth/users/{username} [delete]
func handleDeleteAuthUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireAuthAdmin(w, r)
	if !ok {
		return
	}
	if authUsersDB == nil {
		jsonErr(w, errors.New("local user store unavailable"), http.StatusServiceUnavailable)
		return
	}
	if err := authUsersDB.deleteUser(r.PathValue("username")); err != nil {
		jsonErr(w, err, http.StatusInternalServerError)
		return
	}
	if sink := currentAuditSink(); sink != nil {
		sink(claims, r, http.StatusOK)
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

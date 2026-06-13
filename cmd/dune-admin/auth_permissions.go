package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// The permissions matrix maps Discord role IDs to capability name lists.
// Stored as plain strings (not capability) so unknown entries from an older
// or newer file version degrade gracefully instead of failing the load.
var permissionsState struct {
	mu     sync.RWMutex
	matrix map[string][]string
}

// permissionsPathOverride lets tests redirect persistence; empty → default.
var permissionsPathOverride string

func permissionsPath() string {
	if permissionsPathOverride != "" {
		return permissionsPathOverride
	}
	return filepath.Join(configDir(), "permissions.yaml")
}

func withPermissionsPath(t interface{ Cleanup(func()) }, path string) {
	old := permissionsPathOverride
	permissionsPathOverride = path
	t.Cleanup(func() { permissionsPathOverride = old })
}

func snapshotPermissionsMatrix() map[string][]string {
	permissionsState.mu.RLock()
	defer permissionsState.mu.RUnlock()
	out := make(map[string][]string, len(permissionsState.matrix))
	for k, v := range permissionsState.matrix {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func setPermissionsMatrix(m map[string][]string) {
	permissionsState.mu.Lock()
	defer permissionsState.mu.Unlock()
	permissionsState.matrix = m
}

// applyMatrixRow unions the capabilities mapped to key into caps. Callers
// hold permissionsState.mu (read).
func applyMatrixRow(caps map[capability]bool, key string) {
	for _, name := range permissionsState.matrix[key] {
		cap := capability(name)
		if _, ok := allCapabilities[cap]; ok {
			caps[cap] = true
		}
	}
}

// capsForRoles returns the cascade for a member session: the "default"
// pseudo-row (inherited by every authenticated non-owner) unioned with the
// rows for each of the session's Discord roles. Nothing is granted
// implicitly — an empty matrix means no access.
func capsForRoles(roleIDs []string) map[capability]bool {
	caps := map[capability]bool{}
	permissionsState.mu.RLock()
	defer permissionsState.mu.RUnlock()
	applyMatrixRow(caps, pseudoRoleDefault)
	for _, role := range roleIDs {
		applyMatrixRow(caps, role)
	}
	return caps
}

// capsForSession resolves the capability set for any non-owner session as a
// cascade: EVERY session inherits the "default" pseudo-row, then layers on
// the rows that apply to it — "guest" for guest sessions, the Discord role
// rows for members, plus a local DB user's directly-assigned capabilities.
// No capability is implicit; absence of a grant is denial.
func capsForSession(claims *sessionClaims) map[capability]bool {
	if claims.Method == "guest" {
		caps := map[capability]bool{}
		permissionsState.mu.RLock()
		defer permissionsState.mu.RUnlock()
		applyMatrixRow(caps, pseudoRoleDefault)
		applyMatrixRow(caps, pseudoRoleGuest)
		return caps
	}
	caps := capsForRoles(claims.RoleIDs)
	if claims.Method == "local" && authUsersDB != nil {
		username := strings.TrimPrefix(claims.Sub, "local:")
		for _, name := range authUsersDB.capsForUser(username) {
			cap := capability(name)
			if _, ok := allCapabilities[cap]; ok {
				caps[cap] = true
			}
		}
	}
	return caps
}

// loadPermissionsMatrix reads the matrix file. A missing file is an empty
// matrix (defaults apply); a corrupt file is an error so it is never
// silently overwritten.
func loadPermissionsMatrix(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string][]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read permissions: %w", err)
	}
	var m map[string][]string
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse permissions: %w", err)
	}
	if m == nil {
		m = map[string][]string{}
	}
	return m, nil
}

// savePermissionsMatrix writes the matrix atomically (write temp + rename)
// so a crash mid-write cannot corrupt the file.
func savePermissionsMatrix(path string, m map[string][]string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal permissions: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write permissions: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace permissions: %w", err)
	}
	return nil
}

// defaultSeedCaps is the read-only baseline written to the "default" matrix
// row on first run. It is seeded into the matrix — not hardcoded into the
// resolver — so it is visible and fully editable in the Permissions UI. An
// admin who clears it gets true deny-by-default for everyone.
func defaultSeedCaps() []string {
	return []string{
		string(capPlayersRead),
		string(capWorldRead),
		string(capServerRead),
		string(capMarketRead),
		string(capBattlepassTrack),
	}
}

// initPermissionsMatrix loads the persisted matrix at startup. On a fresh
// install (no permissions file yet) it seeds the "default" row with a
// read-only baseline and persists it, so the cascade has a sensible starting
// point that the operator can then tighten or expand.
func initPermissionsMatrix() {
	path := permissionsPath()
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		seed := map[string][]string{pseudoRoleDefault: defaultSeedCaps()}
		if err := savePermissionsMatrix(path, seed); err != nil {
			logAuthError("seed permissions matrix: " + err.Error())
		}
		setPermissionsMatrix(seed)
		return
	}
	m, err := loadPermissionsMatrix(path)
	if err != nil {
		logAuthError("permissions matrix load failed (deny-by-default applies, file left untouched): " + err.Error())
		return
	}
	setPermissionsMatrix(m)
}

// ── owner / auth:manage endpoints ─────────────────────────────────────────

type capabilityInfo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type permissionsResponse struct {
	Matrix       map[string][]string `json:"matrix"`
	Capabilities []capabilityInfo    `json:"capabilities"`
	GuildRoles   []discordRoleRow    `json:"guild_roles"`
	// SeedDefaults is the standard read-only baseline the "Reset" button in
	// the UI restores the Default row to. Sourced from defaultSeedCaps so the
	// UI never hardcodes it.
	SeedDefaults []string `json:"seed_defaults"`
}

// handleGetPermissions returns the matrix, the capability catalog, the
// default read-only set, and the guild's roles for the editor UI.
//
// @Summary Get the role→capability permissions matrix (owner only)
// @Tags auth
// @Produce json
// @Success 200 {object} permissionsResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/auth/permissions [get]
func handleGetPermissions(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthAdmin(w, r); !ok {
		return
	}
	resp := permissionsResponse{
		Matrix:       snapshotPermissionsMatrix(),
		GuildRoles:   []discordRoleRow{},
		SeedDefaults: defaultSeedCaps(),
	}
	for cap, desc := range allCapabilities {
		resp.Capabilities = append(resp.Capabilities, capabilityInfo{ID: string(cap), Description: desc})
	}
	sort.Slice(resp.Capabilities, func(i, j int) bool {
		return resp.Capabilities[i].ID < resp.Capabilities[j].ID
	})
	// Guild roles are best-effort decoration for the editor; the matrix is
	// editable by raw role ID even when Discord is unavailable. Uses the
	// gateway bot when running, REST-only bot token otherwise.
	if fetch, guildID := discordRolesFetcher(); fetch != nil && guildID != "" {
		if roles, err := cmdListDiscordRoles(guildID, fetch); err == nil {
			resp.GuildRoles = roles
		}
	}
	jsonOK(w, resp)
}

type permissionsUpdateRequest struct {
	Matrix map[string][]string `json:"matrix"`
}

// handlePutPermissions validates, persists, and live-applies a new matrix.
//
// @Summary Replace the permissions matrix (owner only)
// @Tags auth
// @Accept json
// @Produce json
// @Param matrix body permissionsUpdateRequest true "New role→capability matrix"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/auth/permissions [put]
func handlePutPermissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireAuthAdmin(w, r)
	if !ok {
		return
	}
	var req permissionsUpdateRequest
	if err := decode(r, &req); err != nil {
		jsonErr(w, fmt.Errorf("decode: %w", err), http.StatusBadRequest)
		return
	}
	if req.Matrix == nil {
		jsonErr(w, errors.New("matrix is required"), http.StatusBadRequest)
		return
	}
	for role, caps := range req.Matrix {
		for _, name := range caps {
			if _, ok := allCapabilities[capability(name)]; !ok {
				jsonErr(w, fmt.Errorf("unknown capability %q for role %s", name, role), http.StatusBadRequest)
				return
			}
		}
	}
	if err := savePermissionsMatrix(permissionsPath(), req.Matrix); err != nil {
		jsonErr(w, err, http.StatusInternalServerError)
		return
	}
	setPermissionsMatrix(req.Matrix)
	// /api/v1/auth/* bypasses the middleware's audit hook, so record this
	// mutation explicitly — matrix edits are exactly what an audit trail is for.
	if sink := currentAuditSink(); sink != nil {
		sink(claims, r, http.StatusOK)
	}
	jsonOK(w, map[string]string{"status": "saved"})
}

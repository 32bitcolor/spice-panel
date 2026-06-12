package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupBattlepassStore sets globalBattlepassStore to a fresh in-memory store
// and restores nil on cleanup. NOT parallel — mutates package global.
func setupBattlepassStore(t *testing.T) *battlepassStore {
	t.Helper()
	s := testBattlepassStore(t)
	globalBattlepassStore = s
	t.Cleanup(func() { globalBattlepassStore = nil })
	return s
}

// ── nil-guard tests ───────────────────────────────────────────────────────────

func TestBattlepassHandlers_NilStore(t *testing.T) {
	globalBattlepassStore = nil
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		target  string
		pathID  string
	}{
		{"tiers", handleListBattlepassTiers, http.MethodGet, "/api/v1/battlepass/tiers", ""},
		{"update", handleUpdateBattlepassTier, http.MethodPut, "/api/v1/battlepass/tiers/1", "1"},
		{"progress", handleBattlepassProgress, http.MethodGet, "/api/v1/battlepass/progress/1", "1"},
		{"pending", handleBattlepassPending, http.MethodGet, "/api/v1/battlepass/pending", ""},
		{"reseed", handleBattlepassReseed, http.MethodPost, "/api/v1/battlepass/reseed", ""},
		{"grant", handleBattlepassGrant, http.MethodPost, "/api/v1/battlepass/grant", ""},
		{"bulk", handleBattlepassTiersBulk, http.MethodPost, "/api/v1/battlepass/tiers/bulk", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.target, bytes.NewReader([]byte(`{}`)))
			if c.pathID != "" {
				req.SetPathValue("id", c.pathID)
				req.SetPathValue("accountId", c.pathID)
			}
			rec := httptest.NewRecorder()
			c.handler(rec, req)
			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("want 503, got %d", rec.Code)
			}
		})
	}
}

// ── tiers ─────────────────────────────────────────────────────────────────────

func TestHandleListBattlepassTiers(t *testing.T) {
	s := setupBattlepassStore(t)
	if _, err := s.seedTiersIfEmpty(testTiers()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = s.recordClaim("level:5", 1, 10, battlepassClaimEarned)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/battlepass/tiers", nil)
	rec := httptest.NewRecorder()
	handleListBattlepassTiers(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Tiers  []battlepassTier                `json:"tiers"`
		Counts map[string]battlepassTierCounts `json:"counts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Tiers) != 3 {
		t.Fatalf("tiers len = %d, want 3", len(resp.Tiers))
	}
	if resp.Counts["level:5"].Earned != 1 {
		t.Fatalf("counts = %+v, want level:5 earned 1", resp.Counts)
	}
}

func TestHandleUpdateBattlepassTier(t *testing.T) {
	s := setupBattlepassStore(t)
	if _, err := s.seedTiersIfEmpty(testTiers()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	tiers, _ := s.listTiers()

	rewards := `[{"template":"Kindjal_4","qty":1,"quality":3}]`
	body, _ := json.Marshal(map[string]any{"intel": 99, "enabled": false, "label": "Custom", "reward_items": rewards})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/battlepass/tiers/1", bytes.NewReader(body))
	req.SetPathValue("id", fmt.Sprintf("%d", tiers[0].ID))
	rec := httptest.NewRecorder()
	handleUpdateBattlepassTier(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tier battlepassTier
	_ = json.Unmarshal(rec.Body.Bytes(), &tier)
	if tier.Intel != 99 || tier.Enabled || tier.Label != "Custom" || tier.RewardItems != rewards {
		t.Fatalf("updated tier = %+v", tier)
	}

	// Omitted label/reward_items keep their values (inline intel edit path).
	body, _ = json.Marshal(map[string]any{"intel": 50, "enabled": true})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/battlepass/tiers/1", bytes.NewReader(body))
	req.SetPathValue("id", fmt.Sprintf("%d", tiers[0].ID))
	rec = httptest.NewRecorder()
	handleUpdateBattlepassTier(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("partial update: want 200, got %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &tier)
	if tier.Intel != 50 || tier.Label != "Custom" || tier.RewardItems != rewards {
		t.Fatalf("partial update lost fields: %+v", tier)
	}

	// Malformed reward_items rejected.
	bad := "not json"
	body, _ = json.Marshal(map[string]any{"intel": 5, "enabled": true, "reward_items": bad})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/battlepass/tiers/1", bytes.NewReader(body))
	req.SetPathValue("id", fmt.Sprintf("%d", tiers[0].ID))
	rec = httptest.NewRecorder()
	handleUpdateBattlepassTier(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad reward_items: want 400, got %d", rec.Code)
	}
}

func TestHandleBattlepassTiersBulk(t *testing.T) {
	s := setupBattlepassStore(t)
	if _, err := s.seedTiersIfEmpty(testTiers()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	tiers, _ := s.listTiers()
	ids := []int64{tiers[0].ID, tiers[1].ID}

	do := func(action string, ids []int64) *httptest.ResponseRecorder {
		t.Helper()
		body, _ := json.Marshal(map[string]any{"ids": ids, "action": action})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/battlepass/tiers/bulk", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		handleBattlepassTiersBulk(rec, req)
		return rec
	}

	if rec := do("disable", ids); rec.Code != http.StatusOK {
		t.Fatalf("disable: want 200, got %d", rec.Code)
	}
	after, _ := s.listTiers()
	if after[0].Enabled || after[1].Enabled || !after[2].Enabled {
		t.Fatalf("disable did not apply: %+v", after)
	}

	if rec := do("enable", ids); rec.Code != http.StatusOK {
		t.Fatalf("enable: want 200, got %d", rec.Code)
	}
	if rec := do("delete", []int64{tiers[2].ID}); rec.Code != http.StatusOK {
		t.Fatalf("delete: want 200, got %d", rec.Code)
	}
	after, _ = s.listTiers()
	if len(after) != 2 {
		t.Fatalf("after delete %d tiers, want 2", len(after))
	}

	if rec := do("explode", ids); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad action: want 400, got %d", rec.Code)
	}
	if rec := do("enable", nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("no ids: want 400, got %d", rec.Code)
	}
}

func TestHandleUpdateBattlepassTier_BadInput(t *testing.T) {
	setupBattlepassStore(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/battlepass/tiers/x", bytes.NewReader([]byte(`{}`)))
	req.SetPathValue("id", "x")
	rec := httptest.NewRecorder()
	handleUpdateBattlepassTier(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id: want 400, got %d", rec.Code)
	}

	body, _ := json.Marshal(map[string]any{"intel": -5, "enabled": true})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/battlepass/tiers/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	rec = httptest.NewRecorder()
	handleUpdateBattlepassTier(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("negative intel: want 400, got %d", rec.Code)
	}

	body, _ = json.Marshal(map[string]any{"intel": 5, "enabled": true})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/battlepass/tiers/9999", bytes.NewReader(body))
	req.SetPathValue("id", "9999")
	rec = httptest.NewRecorder()
	handleUpdateBattlepassTier(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing tier: want 404, got %d", rec.Code)
	}
}

// ── progress ──────────────────────────────────────────────────────────────────

func TestHandleBattlepassProgress(t *testing.T) {
	s := setupBattlepassStore(t)
	_ = s.recordClaim("level:5", 42, 10, battlepassClaimBaseline)
	_ = s.recordClaim("level:10", 42, 15, battlepassClaimEarned)
	_ = s.recordClaim("level:15", 42, 20, battlepassClaimGranted)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/battlepass/progress/42", nil)
	req.SetPathValue("accountId", "42")
	rec := httptest.NewRecorder()
	handleBattlepassProgress(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Claims       []battlepassClaim `json:"claims"`
		PendingIntel int64             `json:"pending_intel"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Claims) != 3 {
		t.Fatalf("claims len = %d, want 3", len(resp.Claims))
	}
	if resp.PendingIntel != 15 {
		t.Fatalf("pending intel = %d, want 15 (earned only)", resp.PendingIntel)
	}
}

func TestHandleBattlepassProgress_BadID(t *testing.T) {
	setupBattlepassStore(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/battlepass/progress/x", nil)
	req.SetPathValue("accountId", "x")
	rec := httptest.NewRecorder()
	handleBattlepassProgress(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// ── reseed ────────────────────────────────────────────────────────────────────

func TestHandleBattlepassReseed(t *testing.T) {
	s := setupBattlepassStore(t)
	_ = s.recordClaim("level:5", 1, 10, battlepassClaimGranted)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/battlepass/reseed", nil)
	rec := httptest.NewRecorder()
	handleBattlepassReseed(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	tiers, _ := s.listTiers()
	if len(tiers) != len(defaultBattlepassCatalog()) {
		t.Fatalf("tiers = %d, want full default catalog", len(tiers))
	}
	keys, _ := s.claimedKeys(1)
	if keys["level:5"] != battlepassClaimGranted {
		t.Fatal("reseed must preserve claims")
	}
}

// ── grant ─────────────────────────────────────────────────────────────────────

func grantTestDeps(players []battlepassPlayer, awardErr error, awarded *[]int64) battlepassGrantDeps {
	return battlepassGrantDeps{
		fetchPlayers: func(ctx context.Context) ([]battlepassPlayer, error) {
			return players, nil
		},
		awardIntel: func(ctx context.Context, pawnID, amount int64) error {
			if awardErr != nil {
				return awardErr
			}
			if awarded != nil {
				*awarded = append(*awarded, pawnID, amount)
			}
			return nil
		},
		giveItem: func(ctx context.Context, actorID int64, template string, qty, quality int64) error {
			return nil
		},
	}
}

func TestGrantBattlepassEarned_DeliversItemRewards(t *testing.T) {
	s := testBattlepassStore(t)
	rewarded := testTiers()
	rewarded[0].RewardItems = `[{"template":"Kindjal_4","qty":2,"quality":3}]`
	if _, err := s.seedTiersIfEmpty(rewarded); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = s.recordClaim("level:5", 1, 10, battlepassClaimEarned)
	_ = s.recordClaim("journey:DA_MQ_FindTheFremen", 1, 40, battlepassClaimEarned)

	type given struct {
		actorID  int64
		template string
		qty      int64
		quality  int64
	}
	var items []given
	deps := grantTestDeps([]battlepassPlayer{{AccountID: 1, PawnID: 100}}, nil, nil)
	deps.giveItem = func(ctx context.Context, actorID int64, template string, qty, quality int64) error {
		items = append(items, given{actorID, template, qty, quality})
		return nil
	}

	if _, _, err := grantBattlepassEarned(context.Background(), s, deps, 1); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("gave %d items, want 1 (only level:5 has rewards)", len(items))
	}
	if items[0] != (given{100, "Kindjal_4", 2, 3}) {
		t.Fatalf("gave %+v", items[0])
	}
}

func TestGrantBattlepassEarned_Success(t *testing.T) {
	s := testBattlepassStore(t)
	_ = s.recordClaim("level:5", 1, 10, battlepassClaimEarned)
	_ = s.recordClaim("level:10", 1, 15, battlepassClaimEarned)
	_ = s.recordClaim("level:15", 1, 20, battlepassClaimBaseline)

	var awarded []int64
	deps := grantTestDeps([]battlepassPlayer{{AccountID: 1, PawnID: 100, Online: false}}, nil, &awarded)

	total, n, err := grantBattlepassEarned(context.Background(), s, deps, 1)
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	if total != 25 || n != 2 {
		t.Fatalf("granted %d intel over %d tiers, want 25/2", total, n)
	}
	if len(awarded) != 2 || awarded[0] != 100 || awarded[1] != 25 {
		t.Fatalf("awardIntel called with %v, want [100 25]", awarded)
	}
	keys, _ := s.claimedKeys(1)
	if keys["level:5"] != battlepassClaimGranted || keys["level:10"] != battlepassClaimGranted {
		t.Fatalf("claims after grant = %v, want granted", keys)
	}
	if keys["level:15"] != battlepassClaimBaseline {
		t.Fatal("baseline claim must remain baseline")
	}
}

func TestGrantBattlepassEarned_NothingEarned(t *testing.T) {
	s := testBattlepassStore(t)
	deps := grantTestDeps([]battlepassPlayer{{AccountID: 1, PawnID: 100}}, nil, nil)
	if _, _, err := grantBattlepassEarned(context.Background(), s, deps, 1); err != errBattlepassNothingEarned {
		t.Fatalf("err = %v, want errBattlepassNothingEarned", err)
	}
}

func TestGrantBattlepassEarned_UnknownAccount(t *testing.T) {
	s := testBattlepassStore(t)
	_ = s.recordClaim("level:5", 1, 10, battlepassClaimEarned)
	deps := grantTestDeps([]battlepassPlayer{{AccountID: 2, PawnID: 200}}, nil, nil)
	if _, _, err := grantBattlepassEarned(context.Background(), s, deps, 1); err != errNotFound {
		t.Fatalf("err = %v, want errNotFound", err)
	}
}

func TestGrantBattlepassEarned_AwardFailureKeepsEarned(t *testing.T) {
	s := testBattlepassStore(t)
	_ = s.recordClaim("level:5", 1, 10, battlepassClaimEarned)
	deps := grantTestDeps([]battlepassPlayer{{AccountID: 1, PawnID: 100, Online: true}},
		fmt.Errorf("player is currently Online"), nil)

	if _, _, err := grantBattlepassEarned(context.Background(), s, deps, 1); err == nil {
		t.Fatal("expected grant error")
	}
	claims, _ := s.listClaims(1)
	if claims[0].Status != battlepassClaimEarned || claims[0].Attempts != 1 || claims[0].LastError == "" {
		t.Fatalf("claim after failed grant = %+v, want earned with attempt recorded", claims[0])
	}
}

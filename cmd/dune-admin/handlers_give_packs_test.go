package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ── handleGetGivePacksConfig ─────────────────────────────────────────────────

func TestHandleGetGivePacksConfig_NilStore503(t *testing.T) {
	givePacksStoreDB = nil

	req := httptest.NewRequest(http.MethodGet, "/api/v1/give-packs/config", nil)
	rec := httptest.NewRecorder()
	handleGetGivePacksConfig(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleGetGivePacksConfig_EmptyStore(t *testing.T) {
	setupGivePacksStore(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/give-packs/config", nil)
	rec := httptest.NewRecorder()
	handleGetGivePacksConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp givePacksConfigResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Packs == nil {
		t.Error("expected non-nil packs slice (empty), got nil")
	}
	if len(resp.Packs) != 0 {
		t.Errorf("expected empty packs on fresh store, got %d", len(resp.Packs))
	}
}

func TestHandleGetGivePacksConfig_ReturnsSavedPacks(t *testing.T) {
	s := setupGivePacksStore(t)

	packs := []givePack{
		{ID: "starter-t1", Name: "T1", Category: "Starter", Tier: 1, Items: []welcomePackageItem{
			{Template: "Ammo", Qty: 500, Quality: 0},
		}},
	}
	packsJSON, _ := json.Marshal(packs)
	if err := s.saveConfig(string(packsJSON), true); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/give-packs/config", nil)
	rec := httptest.NewRecorder()
	handleGetGivePacksConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp givePacksConfigResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(resp.Packs))
	}
	if resp.Packs[0].ID != "starter-t1" {
		t.Errorf("expected id=starter-t1, got %q", resp.Packs[0].ID)
	}
}

// ── handlePutGivePacksConfig ─────────────────────────────────────────────────

func TestHandlePutGivePacksConfig_NilStore503(t *testing.T) {
	givePacksStoreDB = nil

	body, _ := json.Marshal(givePacksConfigResponse{Packs: []givePack{}})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/give-packs/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handlePutGivePacksConfig(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlePutGivePacksConfig_BadBody400(t *testing.T) {
	setupGivePacksStore(t)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/give-packs/config", bytes.NewReader([]byte("not-json")))
	rec := httptest.NewRecorder()
	handlePutGivePacksConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlePutGivePacksConfig_ValidationError400(t *testing.T) {
	setupGivePacksStore(t)

	// Duplicate id → validation fails.
	badPacks := givePacksConfigResponse{Packs: []givePack{
		{ID: "dup", Name: "A", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: 1, Quality: 0}}},
		{ID: "dup", Name: "B", Category: "Y", Tier: 2, Items: []welcomePackageItem{{Template: "B", Qty: 1, Quality: 0}}},
	}}
	body, _ := json.Marshal(badPacks)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/give-packs/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handlePutGivePacksConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlePutGivePacksConfig_PersistsAndReturns(t *testing.T) {
	setupGivePacksStore(t)

	packs := givePacksConfigResponse{Packs: []givePack{
		{ID: "buggy-t6", Name: "T6", Category: "Buggy", Tier: 6, Items: []welcomePackageItem{
			{Template: "BuggyBoost_6", Qty: 1, Quality: 0},
		}},
	}}
	body, _ := json.Marshal(packs)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/give-packs/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handlePutGivePacksConfig(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp givePacksConfigResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode PUT response: %v", err)
	}
	if len(resp.Packs) != 1 || resp.Packs[0].ID != "buggy-t6" {
		t.Fatalf("unexpected response packs: %+v", resp.Packs)
	}
}

func TestHandlePutGivePacksConfig_RoundTrip(t *testing.T) {
	setupGivePacksStore(t)

	putPacks := givePacksConfigResponse{Packs: []givePack{
		{ID: "scout-t3", Name: "T3", Category: "Scout", Tier: 3, Items: []welcomePackageItem{
			{Template: "ScoutPart_3", Qty: 2, Quality: 0},
		}},
	}}
	body, _ := json.Marshal(putPacks)
	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/give-packs/config", bytes.NewReader(body))
	putRec := httptest.NewRecorder()
	handlePutGivePacksConfig(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT: want 200, got %d: %s", putRec.Code, putRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/give-packs/config", nil)
	getRec := httptest.NewRecorder()
	handleGetGivePacksConfig(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET: want 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	var got givePacksConfigResponse
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatalf("decode GET: %v", err)
	}
	if len(got.Packs) != 1 || got.Packs[0].ID != "scout-t3" {
		t.Fatalf("GET after PUT returned wrong packs: %+v", got.Packs)
	}
}

func TestHandlePutGivePacksConfig_EmptyPacksNoReSeed(t *testing.T) {
	// Deleting all packs (empty PUT) must NOT trigger re-seed on the next GET.
	s := setupGivePacksStore(t)

	// Pre-seed with some data.
	if err := s.saveConfig(`[{"id":"x","name":"X","category":"X","tier":1,"items":[]}]`, true); err != nil {
		t.Fatalf("pre-seed: %v", err)
	}

	// PUT empty packs.
	emptyBody, _ := json.Marshal(givePacksConfigResponse{Packs: []givePack{}})
	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/give-packs/config", bytes.NewReader(emptyBody))
	putRec := httptest.NewRecorder()
	handlePutGivePacksConfig(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT empty: want 200, got %d: %s", putRec.Code, putRec.Body.String())
	}

	// GET must return empty, no re-seed.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/give-packs/config", nil)
	getRec := httptest.NewRecorder()
	handleGetGivePacksConfig(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET: want 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	var got givePacksConfigResponse
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatalf("decode GET: %v", err)
	}
	if len(got.Packs) != 0 {
		t.Fatalf("expected 0 packs after empty PUT, got %d (re-seed must NOT happen)", len(got.Packs))
	}
}

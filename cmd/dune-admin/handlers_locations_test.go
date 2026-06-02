package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupLocationStore sets globalLocationStore to a fresh in-memory store and
// restores nil on cleanup. NOT parallel — mutates package global.
func setupLocationStore(t *testing.T) *locationStore {
	t.Helper()
	s := openMemLocationStore(t)
	globalLocationStore = s
	t.Cleanup(func() { globalLocationStore = nil })
	return s
}

// ── nil-guard tests (globalLocationStore == nil) ─────────────────────────────

func TestHandleListLocations_NilStore(t *testing.T) {
	globalLocationStore = nil
	req := httptest.NewRequest(http.MethodGet, "/api/v1/locations", nil)
	rec := httptest.NewRecorder()
	handleListLocations(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleUpsertLocation_NilStore(t *testing.T) {
	globalLocationStore = nil
	body, _ := json.Marshal(map[string]any{"name": "X", "x": 1.0, "y": 2.0, "z": 3.0})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleUpsertLocation(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleRenameLocation_NilStore(t *testing.T) {
	globalLocationStore = nil
	body, _ := json.Marshal(map[string]string{"old_name": "A", "new_name": "B"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleRenameLocation(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleDeleteLocation_NilStore(t *testing.T) {
	globalLocationStore = nil
	body, _ := json.Marshal(map[string]string{"name": "X"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleDeleteLocation(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

// ── list ──────────────────────────────────────────────────────────────────────

func TestHandleListLocations_ReturnsSeededLocations(t *testing.T) {
	setupLocationStore(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/locations", nil)
	rec := httptest.NewRecorder()
	handleListLocations(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var locs []teleportLocation
	if err := json.Unmarshal(rec.Body.Bytes(), &locs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(locs) != len(cheatLocations) {
		t.Fatalf("want %d locations, got %d", len(cheatLocations), len(locs))
	}
}

// ── upsert ────────────────────────────────────────────────────────────────────

func TestHandleUpsertLocation_AddsNew(t *testing.T) {
	setupLocationStore(t)
	body, _ := json.Marshal(map[string]any{"name": "NewPlace", "x": 1.1, "y": 2.2, "z": 3.3})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleUpsertLocation(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	// Confirm it appears in list.
	locs, _ := globalLocationStore.list()
	var found bool
	for _, l := range locs {
		if l.Name == "NewPlace" {
			found = true
		}
	}
	if !found {
		t.Fatal("upserted location not in store")
	}
}

func TestHandleUpsertLocation_RejectsMissingName(t *testing.T) {
	setupLocationStore(t)
	body, _ := json.Marshal(map[string]any{"x": 1.0, "y": 2.0, "z": 3.0})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleUpsertLocation(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestHandleUpsertLocation_RejectsBadJSON(t *testing.T) {
	setupLocationStore(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/locations", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	handleUpsertLocation(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// ── rename ────────────────────────────────────────────────────────────────────

func TestHandleRenameLocation_Success(t *testing.T) {
	setupLocationStore(t)
	body, _ := json.Marshal(map[string]string{"old_name": "Windsack", "new_name": "Windsack2"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleRenameLocation(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	locs, _ := globalLocationStore.list()
	var found bool
	for _, l := range locs {
		if l.Name == "Windsack2" {
			found = true
		}
	}
	if !found {
		t.Fatal("renamed location not found")
	}
}

func TestHandleRenameLocation_RejectsMissingFields(t *testing.T) {
	setupLocationStore(t)
	tests := []struct {
		name string
		body map[string]string
	}{
		{"missing old_name", map[string]string{"new_name": "B"}},
		{"missing new_name", map[string]string{"old_name": "Windsack"}},
		{"both empty", map[string]string{"old_name": "", "new_name": ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/locations", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			handleRenameLocation(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d", rec.Code)
			}
		})
	}
}

func TestHandleRenameLocation_UnknownNameReturns404(t *testing.T) {
	setupLocationStore(t)
	body, _ := json.Marshal(map[string]string{"old_name": "NoSuch", "new_name": "Else"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleRenameLocation(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestHandleDeleteLocation_Success(t *testing.T) {
	setupLocationStore(t)
	body, _ := json.Marshal(map[string]string{"name": "Windsack"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleDeleteLocation(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	locs, _ := globalLocationStore.list()
	for _, l := range locs {
		if l.Name == "Windsack" {
			t.Fatal("deleted location still present")
		}
	}
}

func TestHandleDeleteLocation_RejectsMissingName(t *testing.T) {
	setupLocationStore(t)
	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleDeleteLocation(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestHandleDeleteLocation_UnknownNameReturns404(t *testing.T) {
	setupLocationStore(t)
	body, _ := json.Marshal(map[string]string{"name": "NoSuch"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/locations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleDeleteLocation(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

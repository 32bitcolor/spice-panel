package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupEventStore sets globalEventStore to a fresh in-memory store and restores
// nil on cleanup. NOT parallel — mutates package global.
func setupEventStore(t *testing.T) *eventStore {
	t.Helper()
	s := openMemEventStore(t)
	globalEventStore = s
	t.Cleanup(func() { globalEventStore = nil })
	return s
}

// ── nil-guard tests ───────────────────────────────────────────────────────────

func TestHandleListEvents_NilStore(t *testing.T) {
	globalEventStore = nil
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handleListEvents(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleCreateEvent_NilStore(t *testing.T) {
	globalEventStore = nil
	body, _ := json.Marshal(map[string]any{"name": "X", "type": "zone_race"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleCreateEvent(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleUpdateEvent_NilStore(t *testing.T) {
	globalEventStore = nil
	body, _ := json.Marshal(map[string]any{"name": "X"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handleUpdateEvent(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleDeleteEvent_NilStore(t *testing.T) {
	globalEventStore = nil
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handleDeleteEvent(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleSetEventEnabled_NilStore(t *testing.T) {
	globalEventStore = nil
	body, _ := json.Marshal(map[string]bool{"enabled": true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/1/enable", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handleSetEventEnabled(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleGetEventStatus_NilStore(t *testing.T) {
	globalEventStore = nil
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/1/status", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handleGetEventStatus(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

func TestHandleResetEvent_NilStore(t *testing.T) {
	globalEventStore = nil
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/1/reset", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()
	handleResetEvent(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
}

// ── list ──────────────────────────────────────────────────────────────────────

func TestHandleListEvents_Empty(t *testing.T) {
	setupEventStore(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handleListEvents(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var result []eventDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("want empty list, got %d items", len(result))
	}
}

func TestHandleListEvents_ReturnsAll(t *testing.T) {
	s := setupEventStore(t)
	mustCreateEvent(t, s, "event_a", eventTypeZoneRace)
	mustCreateEvent(t, s, "event_b", eventTypeMilestone)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handleListEvents(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var result []eventDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("want 2 events, got %d", len(result))
	}
}

// ── create ────────────────────────────────────────────────────────────────────

func TestHandleCreateEvent_ValidInput(t *testing.T) {
	setupEventStore(t)
	body, _ := json.Marshal(map[string]any{
		"name":                "my_race",
		"type":                "zone_race",
		"config":              `{"map":"TestMap","x":0,"y":0,"z":0,"radius":10}`,
		"announce_template":   "{player} wins!",
		"announce_channel_id": "12345",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleCreateEvent(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var result eventDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.ID == 0 {
		t.Fatal("want non-zero ID")
	}
	if result.Name != "my_race" {
		t.Fatalf("want name %q, got %q", "my_race", result.Name)
	}
}

func TestHandleCreateEvent_MissingName(t *testing.T) {
	setupEventStore(t)
	body, _ := json.Marshal(map[string]any{"type": "zone_race"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleCreateEvent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestHandleCreateEvent_InvalidType(t *testing.T) {
	setupEventStore(t)
	body, _ := json.Marshal(map[string]any{"name": "x", "type": "unknown_type"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleCreateEvent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// ── update ────────────────────────────────────────────────────────────────────

func TestHandleUpdateEvent_ValidInput(t *testing.T) {
	s := setupEventStore(t)
	def := mustCreateEvent(t, s, "original", eventTypeZoneRace)

	body, _ := json.Marshal(map[string]any{
		"name": "updated",
		"type": "zone_race",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/1", bytes.NewReader(body))
	req.SetPathValue("id", itoa(def.ID))
	rec := httptest.NewRecorder()
	handleUpdateEvent(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var result eventDefinition
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Name != "updated" {
		t.Fatalf("want name %q, got %q", "updated", result.Name)
	}
}

func TestHandleUpdateEvent_NotFound(t *testing.T) {
	setupEventStore(t)
	body, _ := json.Marshal(map[string]any{"name": "x", "type": "zone_race"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/999", bytes.NewReader(body))
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	handleUpdateEvent(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestHandleUpdateEvent_BadID(t *testing.T) {
	setupEventStore(t)
	body, _ := json.Marshal(map[string]any{"name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/abc", bytes.NewReader(body))
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()
	handleUpdateEvent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestHandleUpdateEvent_MissingType(t *testing.T) {
	s := setupEventStore(t)
	def := mustCreateEvent(t, s, "original", eventTypeZoneRace)
	body, _ := json.Marshal(map[string]any{"name": "updated"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/1", bytes.NewReader(body))
	req.SetPathValue("id", itoa(def.ID))
	rec := httptest.NewRecorder()
	handleUpdateEvent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestHandleDeleteEvent_ValidInput(t *testing.T) {
	s := setupEventStore(t)
	def := mustCreateEvent(t, s, "to_delete", eventTypeZoneRace)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/1", nil)
	req.SetPathValue("id", itoa(def.ID))
	rec := httptest.NewRecorder()
	handleDeleteEvent(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	// Confirm deleted
	events, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 0 {
		t.Fatal("want empty list after delete")
	}
}

func TestHandleDeleteEvent_NotFound(t *testing.T) {
	setupEventStore(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/events/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	handleDeleteEvent(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

// ── enable / disable ──────────────────────────────────────────────────────────

func TestHandleSetEventEnabled_Enable(t *testing.T) {
	s := setupEventStore(t)
	def := mustCreateEvent(t, s, "toggleme", eventTypeZoneRace)

	body, _ := json.Marshal(map[string]bool{"enabled": true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/1/enable", bytes.NewReader(body))
	req.SetPathValue("id", itoa(def.ID))
	rec := httptest.NewRecorder()
	handleSetEventEnabled(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	updated, err := s.get(def.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !updated.Enabled {
		t.Fatal("want event enabled")
	}
}

func TestHandleSetEventEnabled_NotFound(t *testing.T) {
	setupEventStore(t)
	body, _ := json.Marshal(map[string]bool{"enabled": true})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/999/enable", bytes.NewReader(body))
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	handleSetEventEnabled(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

// ── status ────────────────────────────────────────────────────────────────────

func TestHandleGetEventStatus_ReturnsClaims(t *testing.T) {
	s := setupEventStore(t)
	def := mustCreateEvent(t, s, "statusevent", eventTypeZoneRace)
	if err := s.recordGranted(def.ID, def.Version, 101); err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/1/status", nil)
	req.SetPathValue("id", itoa(def.ID))
	rec := httptest.NewRecorder()
	handleGetEventStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var result struct {
		Event  eventDefinition    `json:"event"`
		Claims []eventClaimRecord `json:"claims"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Claims) != 1 {
		t.Fatalf("want 1 claim, got %d", len(result.Claims))
	}
}

func TestHandleGetEventStatus_NotFound(t *testing.T) {
	setupEventStore(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/999/status", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	handleGetEventStatus(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

// ── reset ─────────────────────────────────────────────────────────────────────

func TestHandleResetEvent_ClearsClaims(t *testing.T) {
	s := setupEventStore(t)
	def := mustCreateEvent(t, s, "resetevent", eventTypeZoneRace)
	if err := s.recordGranted(def.ID, def.Version, 101); err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/1/reset", nil)
	req.SetPathValue("id", itoa(def.ID))
	rec := httptest.NewRecorder()
	handleResetEvent(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	claims, err := s.listClaims(def.ID)
	if err != nil {
		t.Fatalf("listClaims: %v", err)
	}
	if len(claims) != 0 {
		t.Fatalf("want no claims after reset, got %d", len(claims))
	}
}

func TestHandleResetEvent_NotFound(t *testing.T) {
	setupEventStore(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/999/reset", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	handleResetEvent(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

// itoa converts an int64 to its decimal string — avoids importing strconv.
func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// ── events config ─────────────────────────────────────────────────────────────

func TestHandleGetEventsConfig(t *testing.T) {
	orig := loadedConfig
	t.Cleanup(func() { loadedConfig = orig })

	enabled := true
	loadedConfig = appConfig{EventsEnabled: &enabled}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/config", nil)
	rec := httptest.NewRecorder()
	handleGetEventsConfig(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got eventsConfigPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Enabled == nil || !*got.Enabled {
		t.Errorf("Enabled = %v, want true", got.Enabled)
	}
}

func TestHandleSaveEventsConfig(t *testing.T) {
	orig := loadedConfig
	t.Cleanup(func() { loadedConfig = orig })
	t.Setenv("DUNE_ADMIN_CONFIG_DIR", t.TempDir())

	body, _ := json.Marshal(eventsConfigPayload{Enabled: boolPtr(true)})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleSaveEventsConfig(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got eventsConfigPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Enabled == nil || !*got.Enabled {
		t.Errorf("Enabled = %v, want true", got.Enabled)
	}
	if loadedConfig.EventsEnabled == nil || !*loadedConfig.EventsEnabled {
		t.Errorf("loadedConfig.EventsEnabled = %v, want true", loadedConfig.EventsEnabled)
	}
}

func TestHandleSaveEventsConfig_BadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/v1/events/config", bytes.NewReader([]byte(`{bad`)))
	rec := httptest.NewRecorder()
	handleSaveEventsConfig(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestApplyEventEngine_StopsRunningEngine(t *testing.T) {
	// Arrange: place a sentinel cancel in the global so we can detect it gets called.
	called := false
	globalEventsMu.Lock()
	globalEventsCancel = func() { called = true }
	globalEventsMu.Unlock()
	t.Cleanup(func() {
		globalEventsMu.Lock()
		globalEventsCancel = nil
		globalEventsMu.Unlock()
	})

	// Act: apply with enabled=false — no store, so engine won't start.
	disabled := false
	applyEventEngine(appConfig{EventsEnabled: &disabled})

	if !called {
		t.Error("expected existing cancel to be called when engine is stopped")
	}
	globalEventsMu.Lock()
	nilAfter := globalEventsCancel == nil
	globalEventsMu.Unlock()
	if !nilAfter {
		t.Error("expected globalEventsCancel to be nil after stop")
	}
}

func TestApplyEventEngine_NoopWhenDisabledAndStopped(t *testing.T) {
	globalEventsMu.Lock()
	globalEventsCancel = nil
	globalEventsMu.Unlock()

	disabled := false
	// Should not panic or set a new cancel when already stopped.
	applyEventEngine(appConfig{EventsEnabled: &disabled})

	globalEventsMu.Lock()
	c := globalEventsCancel
	globalEventsMu.Unlock()
	if c != nil {
		t.Error("expected globalEventsCancel to remain nil")
	}
}

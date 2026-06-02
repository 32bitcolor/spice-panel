package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestProcessFillWater exercises the online/offline branching with injected
// deps — no DB or broker needed.
func TestProcessFillWater(t *testing.T) {
	t.Parallel()

	t.Run("online player: RMQ path called", func(t *testing.T) {
		t.Parallel()
		var gotFlsID string
		var gotAmount int
		err := processFillWater(fillWaterParams{
			flsID:        "abc123",
			waterAmount:  500000,
			isOnline:     func(string) bool { return true },
			sendRMQ:      func(id string, amt int) error { gotFlsID = id; gotAmount = amt; return nil },
			resolveActor: func(string) (int64, error) { t.Error("resolveActor must not be called online"); return 0, nil },
			refillDB:     func(int64) (int64, error) { t.Error("refillDB must not be called online"); return 0, nil },
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotFlsID != "abc123" || gotAmount != 500000 {
			t.Fatalf("RMQ called with wrong args: flsID=%q amount=%d", gotFlsID, gotAmount)
		}
	})

	t.Run("offline player: DB path called", func(t *testing.T) {
		t.Parallel()
		rmqCalled := false
		var gotActorID int64
		err := processFillWater(fillWaterParams{
			flsID:        "abc123",
			waterAmount:  1000000,
			isOnline:     func(string) bool { return false },
			sendRMQ:      func(string, int) error { rmqCalled = true; return nil },
			resolveActor: func(string) (int64, error) { return 42, nil },
			refillDB:     func(id int64) (int64, error) { gotActorID = id; return 3, nil },
		})
		if rmqCalled {
			t.Fatal("RMQ must not be called for offline player")
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotActorID != 42 {
			t.Fatalf("refillDB called with wrong actor ID: %d", gotActorID)
		}
	})

	t.Run("offline: resolve actor error propagates", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("player not found")
		err := processFillWater(fillWaterParams{
			flsID:        "abc123",
			waterAmount:  1000000,
			isOnline:     func(string) bool { return false },
			sendRMQ:      func(string, int) error { return nil },
			resolveActor: func(string) (int64, error) { return 0, boom },
			refillDB:     func(int64) (int64, error) { return 0, nil },
		})
		if !errors.Is(err, boom) {
			t.Fatalf("want boom, got %v", err)
		}
	})

	t.Run("offline: DB error propagates", func(t *testing.T) {
		t.Parallel()
		boom := errors.New("db down")
		err := processFillWater(fillWaterParams{
			flsID:        "abc123",
			waterAmount:  1000000,
			isOnline:     func(string) bool { return false },
			sendRMQ:      func(string, int) error { return nil },
			resolveActor: func(string) (int64, error) { return 42, nil },
			refillDB:     func(int64) (int64, error) { return 0, boom },
		})
		if !errors.Is(err, boom) {
			t.Fatalf("want boom, got %v", err)
		}
	})

	t.Run("RMQ error propagates", func(t *testing.T) {
		t.Parallel()
		err := processFillWater(fillWaterParams{
			flsID:        "abc123",
			waterAmount:  1000000,
			isOnline:     func(string) bool { return true },
			sendRMQ:      func(string, int) error { return errTestSentinel },
			resolveActor: func(string) (int64, error) { return 0, nil },
			refillDB:     func(int64) (int64, error) { return 0, nil },
		})
		if err == nil {
			t.Fatal("expected RMQ error, got nil")
		}
	})

	t.Run("water amount defaults to 1000000 when zero", func(t *testing.T) {
		t.Parallel()
		var gotAmount int
		_ = processFillWater(fillWaterParams{
			flsID:        "abc123",
			waterAmount:  0,
			isOnline:     func(string) bool { return true },
			sendRMQ:      func(_ string, amt int) error { gotAmount = amt; return nil },
			resolveActor: func(string) (int64, error) { return 0, nil },
			refillDB:     func(int64) (int64, error) { return 0, nil },
		})
		if gotAmount != 1000000 {
			t.Fatalf("want default 1000000, got %d", gotAmount)
		}
	})
}

// TestHandleFillWater_InputValidation verifies bad input returns 400.
func TestHandleFillWater_InputValidation(t *testing.T) {
	tests := []struct {
		name       string
		rawBody    []byte
		wantStatus int
	}{
		{
			name:       "missing fls_id",
			rawBody:    []byte(`{"water_amount":1000}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty fls_id",
			rawBody:    []byte(`{"fls_id":""}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad json",
			rawBody:    []byte(`{bad`),
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/players/fill-water", bytes.NewReader(tt.rawBody))
			rec := httptest.NewRecorder()
			handleRMQFillWater(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d (body: %s)", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

// TestHandleFillWater_OfflineUsesDB verifies that an offline player triggers
// the DB path. globalDB is nil in unit tests, so the DB call returns an error
// and the handler returns 500 (not 422 or the old silent 200).
func TestHandleFillWater_OfflineUsesDB(t *testing.T) {
	// NOT parallel — reads globalDB package global (nil in tests).
	// isHexIDOnline → false (nil DB), so the offline DB path fires.
	// cmdRefillWaterOffline sees globalDB nil and returns an error → 500.
	body, _ := json.Marshal(map[string]any{"fls_id": "abc123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/players/fill-water", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleRMQFillWater(rec, req)
	// 500: DB path attempted but globalDB is nil.
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500 (DB path, nil DB), got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestWaterFillableTemplates verifies the generated list is non-empty and
// contains expected canonical entries from DT_ItemTableFillables.
func TestWaterFillableTemplates(t *testing.T) {
	t.Parallel()
	if len(waterFillableTemplates) == 0 {
		t.Fatal("waterFillableTemplates must not be empty")
	}
	want := []string{"literjon", "decajon", "dewpack", "literjon_t6"}
	set := make(map[string]bool, len(waterFillableTemplates))
	for _, s := range waterFillableTemplates {
		set[s] = true
	}
	for _, w := range want {
		if !set[w] {
			t.Errorf("waterFillableTemplates missing expected entry %q", w)
		}
	}
	// Must not contain blood containers.
	blood := []string{"bloodsack_01", "bloodsack_02", "bloodsack_t6"}
	for _, b := range blood {
		if set[b] {
			t.Errorf("waterFillableTemplates must not contain blood container %q", b)
		}
	}
}

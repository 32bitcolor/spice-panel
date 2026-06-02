package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// errTestSentinel is a shared sentinel used in handler tests that inject errors.
var errTestSentinel = errors.New("injected test error")

// TestProcessTeleportCoords exercises the online/offline branching logic with
// injected deps (no DB or broker), mirroring the processWhisper pattern.
func TestProcessTeleportCoords(t *testing.T) {
	t.Parallel()

	type result struct {
		path    string
		flsID   string
		x, y, z float64
	}

	t.Run("online player uses RMQ path", func(t *testing.T) {
		t.Parallel()
		var got result
		err := processTeleportCoords(teleportCoordsParams{
			flsID: "abc123",
			x:     100, y: 200, z: 300,
			isOnline: func(_ string) bool { return true },
			sendRMQ:  func(id string, x, y, z float64) error { got = result{"rmq", id, x, y, z}; return nil },
			writeDB: func(id string, pid int64, x, y, z float64) error {
				t.Error("DB path must not be called for online player")
				return nil
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.path != "rmq" || got.flsID != "abc123" || got.x != 100 || got.y != 200 || got.z != 300 {
			t.Fatalf("wrong result: %+v", got)
		}
	})

	t.Run("offline player uses DB path", func(t *testing.T) {
		t.Parallel()
		var got result
		err := processTeleportCoords(teleportCoordsParams{
			flsID: "abc123",
			x:     100, y: 200, z: 300,
			partitionID: 7,
			isOnline:    func(_ string) bool { return false },
			sendRMQ: func(id string, x, y, z float64) error {
				t.Error("RMQ must not be called for offline player")
				return nil
			},
			writeDB: func(id string, pid int64, x, y, z float64) error { got = result{"db", id, x, y, z}; return nil },
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.path != "db" || got.flsID != "abc123" {
			t.Fatalf("wrong result: %+v", got)
		}
	})

	t.Run("RMQ error propagates", func(t *testing.T) {
		t.Parallel()
		boom := errTestSentinel
		err := processTeleportCoords(teleportCoordsParams{
			flsID: "abc123",
			x:     1, y: 2, z: 3,
			isOnline: func(_ string) bool { return true },
			sendRMQ:  func(string, float64, float64, float64) error { return boom },
			writeDB:  func(string, int64, float64, float64, float64) error { return nil },
		})
		if err == nil {
			t.Fatal("expected error from RMQ, got nil")
		}
	})

	t.Run("DB error propagates", func(t *testing.T) {
		t.Parallel()
		boom := errTestSentinel
		err := processTeleportCoords(teleportCoordsParams{
			flsID: "abc123",
			x:     1, y: 2, z: 3,
			isOnline: func(_ string) bool { return false },
			sendRMQ:  func(string, float64, float64, float64) error { return nil },
			writeDB:  func(string, int64, float64, float64, float64) error { return boom },
		})
		if err == nil {
			t.Fatal("expected error from DB, got nil")
		}
	})
}

// TestHandleTeleportCoords_InputValidation checks bad input before any
// DB/RMQ call. globalDB is nil so the 503 guard fires for a connected path;
// this only tests the 400 paths which fire first.
func TestHandleTeleportCoords_InputValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]any
		wantStatus int
	}{
		{
			name:       "missing fls_id",
			body:       map[string]any{"x": 1.0, "y": 2.0, "z": 3.0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty fls_id",
			body:       map[string]any{"fls_id": "", "x": 1.0, "y": 2.0, "z": 3.0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad json",
			body:       nil, // signals raw bad JSON below
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body == nil {
				bodyBytes = []byte("{bad")
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/players/teleport-coords", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()
			handleTeleportCoords(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d (body: %s)", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

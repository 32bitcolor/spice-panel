package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// setupWelcomeStore wires a fresh in-memory store into welcomeStoreDB and
// restores nil on cleanup. NOT parallel — mutates a package global.
func setupWelcomeStore(t *testing.T) *welcomeStore {
	t.Helper()
	s := openMemWelcomeStore(t)
	welcomeStoreDB = s
	t.Cleanup(func() { welcomeStoreDB = nil })
	return s
}

// TestHandlePutWelcomeConfig_MessageFieldsPersisted verifies that the welcome
// message fields survive a PUT → GET round-trip via the SQLite store.
// This is a regression test for the bug where the handler built the
// welcomeConfigRow without WelcomeMessage* fields, so they were lost on refresh.
func TestHandlePutWelcomeConfig_MessageFieldsPersisted(t *testing.T) {
	setupWelcomeStore(t)

	payload := welcomeConfigResponse{
		Enabled:                    false,
		ScanIntervalSecs:           30,
		ActiveVersion:              "",
		Packages:                   []welcomePackage{},
		WelcomeMessageEnabled:      true,
		WelcomeMessage:             "Welcome to the server! Enjoy your starter pack.",
		WelcomeWhisperSourcePlayer: "fls-id-abc123",
	}
	body, _ := json.Marshal(payload)

	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/welcome-package/config", bytes.NewReader(body))
	putRec := httptest.NewRecorder()
	handlePutWelcomeConfig(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT: want 200, got %d: %s", putRec.Code, putRec.Body.String())
	}

	// Simulate a UI refresh: GET re-reads the store.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/welcome-package/config", nil)
	getRec := httptest.NewRecorder()
	handleGetWelcomeConfig(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET: want 200, got %d: %s", getRec.Code, getRec.Body.String())
	}

	var got welcomeConfigResponse
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if !got.WelcomeMessageEnabled {
		t.Error("WelcomeMessageEnabled: want true, got false after refresh")
	}
	if got.WelcomeMessage != payload.WelcomeMessage {
		t.Errorf("WelcomeMessage: want %q, got %q", payload.WelcomeMessage, got.WelcomeMessage)
	}
	if got.WelcomeWhisperSourcePlayer != payload.WelcomeWhisperSourcePlayer {
		t.Errorf("WelcomeWhisperSourcePlayer: want %q, got %q", payload.WelcomeWhisperSourcePlayer, got.WelcomeWhisperSourcePlayer)
	}
}

func TestBuildWelcomeRuntime(t *testing.T) {
	t.Parallel()
	pkgs := []welcomePackage{{Version: "v1"}, {Version: "v2"}}
	tests := []struct {
		name         string
		enabled      bool
		active       string
		scanSecs     int
		packages     []welcomePackage
		wantActive   string
		wantInterval time.Duration
	}{
		{"defaults active to first package", true, "", 0, pkgs, "v1", welcomeDefaultScanInterval},
		{"unknown active falls back to first", true, "vX", 0, pkgs, "v1", welcomeDefaultScanInterval},
		{"explicit active respected", true, "v2", 120, pkgs, "v2", 120 * time.Second},
		{"interval below floor is clamped", false, "v1", 1, pkgs, "v1", welcomeDefaultScanInterval},
		{"min interval honored", true, "v2", 5, pkgs, "v2", 5 * time.Second},
		{"no packages → empty active", true, "", 60, nil, "", 60 * time.Second},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rt := buildWelcomeRuntime(tt.enabled, tt.active, tt.scanSecs, tt.packages, welcomeMessageOptions{})
			if rt.enabled != tt.enabled {
				t.Fatalf("enabled: want %v, got %v", tt.enabled, rt.enabled)
			}
			if rt.activeVersion != tt.wantActive {
				t.Fatalf("activeVersion: want %q, got %q", tt.wantActive, rt.activeVersion)
			}
			if rt.interval != tt.wantInterval {
				t.Fatalf("interval: want %v, got %v", tt.wantInterval, rt.interval)
			}
		})
	}
}

func TestWelcomeRuntimeActive(t *testing.T) {
	t.Parallel()
	rt := buildWelcomeRuntime(true, "v2", 30, []welcomePackage{
		{Version: "v1", Items: []welcomePackageItem{{Template: "A", Qty: 1}}},
		{Version: "v2", Items: []welcomePackageItem{{Template: "B", Qty: 2}}},
	}, welcomeMessageOptions{})
	p, ok := rt.active()
	if !ok {
		t.Fatal("expected an active package")
	}
	if p.Version != "v2" || len(p.Items) != 1 || p.Items[0].Template != "B" {
		t.Fatalf("active package wrong: %+v", p)
	}

	empty := buildWelcomeRuntime(true, "", 30, nil, welcomeMessageOptions{})
	if _, ok := empty.active(); ok {
		t.Fatal("expected no active package when library is empty")
	}
}

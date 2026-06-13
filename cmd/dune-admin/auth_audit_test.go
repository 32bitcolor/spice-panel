package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenAuditLog(t *testing.T) {
	t.Run("creates append-only file with 0600", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "audit.log")
		logger, closeFn, err := openAuditLog(path)
		if err != nil {
			t.Fatalf("openAuditLog: %v", err)
		}
		defer closeFn()
		logger.Info("test entry", "user", "local:admin")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("mode = %v, want 0600", info.Mode().Perm())
		}
	})

	t.Run("appends across opens", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "audit.log")
		for range 2 {
			logger, closeFn, err := openAuditLog(path)
			if err != nil {
				t.Fatal(err)
			}
			logger.Info("entry")
			closeFn()
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Count(strings.TrimSpace(string(data)), "\n") + 1
		if lines != 2 {
			t.Errorf("lines = %d, want 2 (append, not truncate)", lines)
		}
	})
}

func TestAuditMiddlewareIntegration(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	enabled := true
	withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, secret)

	path := filepath.Join(t.TempDir(), "audit.log")
	restore, err := installAuditSink(path)
	if err != nil {
		t.Fatalf("installAuditSink: %v", err)
	}
	t.Cleanup(restore)

	mux := authTestMux(t)
	handler := authMiddleware(mux, mux)
	ownerTok := mintTestSession(t, sessionClaims{Sub: "local:admin", Name: "admin", Method: "local", Owner: true}, secret)

	// One read (not audited) + one mutation (audited).
	doAuthRequest(handler, "GET", "/api/v1/players", ownerTok)
	doAuthRequest(handler, "POST", "/api/v1/players/give-item", ownerTok)

	// Wait for the write to land (synchronous, but be safe on slow disks).
	deadline := time.Now().Add(2 * time.Second)
	var data []byte
	for time.Now().Before(deadline) {
		data, _ = os.ReadFile(path)
		if len(data) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("audit lines = %d, want exactly 1 (mutations only): %q", len(lines), string(data))
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("audit line not JSON: %v", err)
	}
	if entry["user"] != "local:admin" {
		t.Errorf("user = %v", entry["user"])
	}
	if entry["method"] != "POST" {
		t.Errorf("method = %v", entry["method"])
	}
	if entry["path"] != "/api/v1/players/give-item" {
		t.Errorf("path = %v", entry["path"])
	}
	if entry["status"] != float64(http.StatusOK) {
		t.Errorf("status = %v", entry["status"])
	}
}

func TestAuditDisabledByDefault(t *testing.T) {
	// With no sink installed, mutations must flow through untouched.
	secret := []byte("0123456789abcdef0123456789abcdef")
	enabled := true
	withAuthTestConfig(t, appConfig{AuthEnabled: &enabled}, secret)
	if currentAuditSink() != nil {
		t.Fatal("audit sink non-nil at test start")
	}
	mux := authTestMux(t)
	handler := authMiddleware(mux, mux)
	ownerTok := mintTestSession(t, sessionClaims{Sub: "local:admin", Method: "local", Owner: true}, secret)
	w := doAuthRequest(handler, "POST", "/api/v1/players/give-item", ownerTok)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}

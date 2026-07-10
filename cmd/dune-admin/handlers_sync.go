package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// ── In-panel "sync from upstream" ────────────────────────────────────────────
// A fork-specific alternative to the binary-swap self-update. Rather than
// replacing the running binary with a vanilla upstream release (which would
// wipe the spice-panel UI), this runs scripts/sync-upstream.sh on the host: it
// merges the latest upstream release (backend + a UI-agnostic frontend
// allowlist), keeps the entire spice-panel frontend, rebuilds the embedded
// binary, swaps it in, then re-execs into it (reusing scheduleRestart).
//
// The script emits `STEP <n> <msg>` / `INFO <msg>` progress lines which are
// polled by the UI via GET /update/sync/status. Script exit code 3 means
// "already up to date" (no-op); any other non-zero is a failure (the script
// restores the tree and never swaps a bad binary).

const syncTotalSteps = 7

// syncScriptPath returns the on-disk path to the sync engine (env-overridable).
func syncScriptPath() string {
	if p := os.Getenv("SPICE_SYNC_SCRIPT"); p != "" {
		return p
	}
	return "/home/ladmin/spice-panel/scripts/sync-upstream.sh"
}

type syncStatus struct {
	Running bool   `json:"running"`
	Step    string `json:"step"` // e.g. "4/7"
	Message string `json:"message"`
	Done    bool   `json:"done"`
	NoOp    bool   `json:"no_op"` // already up to date
	Error   string `json:"error,omitempty"`
}

var (
	syncMu  sync.Mutex
	syncCur syncStatus
)

func updateSyncState(f func(*syncStatus)) {
	syncMu.Lock()
	defer syncMu.Unlock()
	f(&syncCur)
}

// handleUpdateSyncStatus reports the current/last sync progress for UI polling.
func handleUpdateSyncStatus(w http.ResponseWriter, _ *http.Request) {
	syncMu.Lock()
	st := syncCur
	syncMu.Unlock()
	jsonOK(w, st)
}

// handleUpdateSync kicks off a background upstream sync and returns immediately.
func handleUpdateSync(w http.ResponseWriter, _ *http.Request) {
	exe, err := os.Executable()
	if err != nil {
		jsonErr(w, fmt.Errorf("cannot determine executable path"), http.StatusInternalServerError)
		return
	}
	script := syncScriptPath()
	if _, err := os.Stat(script); err != nil {
		jsonErr(w, fmt.Errorf("sync script not found at %s", script), http.StatusServiceUnavailable)
		return
	}

	syncMu.Lock()
	if syncCur.Running {
		syncMu.Unlock()
		jsonErr(w, fmt.Errorf("a sync is already running"), http.StatusConflict)
		return
	}
	syncCur = syncStatus{Running: true, Step: "1/" + strconv.Itoa(syncTotalSteps), Message: "Starting…"}
	syncMu.Unlock()

	go runUpstreamSync(script, exe)
	jsonOK(w, map[string]string{"status": "started"})
}

func runUpstreamSync(script, exe string) {
	// #nosec G204 -- fixed, admin-shipped script path; no user-controlled input.
	cmd := exec.Command("/bin/bash", script)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		finishSync(func(s *syncStatus) { s.Error = "pipe: " + err.Error() })
		return
	}
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Start(); err != nil {
		finishSync(func(s *syncStatus) { s.Error = "start: " + err.Error() })
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		parseSyncLine(scanner.Text())
	}

	switch err = cmd.Wait(); {
	case err == nil:
		finishSync(func(s *syncStatus) { s.Message = "Synced — restarting…" })
		componentLog("sync").Info().Msg("upstream sync complete; re-execing into new binary")
		go scheduleRestart(exe)
	case exitCodeIs(err, 3):
		finishSync(func(s *syncStatus) { s.NoOp = true; s.Message = "Already up to date" })
	default:
		tail := lastLines(errBuf.String(), 6)
		finishSync(func(s *syncStatus) { s.Error = "sync failed: " + tail })
		componentLog("sync").Error().Str("detail", tail).Msg("upstream sync failed")
	}
}

func finishSync(f func(*syncStatus)) {
	updateSyncState(func(s *syncStatus) {
		s.Running = false
		s.Done = true
		f(s)
	})
}

func parseSyncLine(line string) {
	switch {
	case strings.HasPrefix(line, "STEP "):
		num, msg, _ := strings.Cut(strings.TrimPrefix(line, "STEP "), " ")
		updateSyncState(func(s *syncStatus) {
			s.Step = num + "/" + strconv.Itoa(syncTotalSteps)
			s.Message = msg
		})
	case strings.HasPrefix(line, "INFO "):
		msg := strings.TrimPrefix(line, "INFO ")
		updateSyncState(func(s *syncStatus) { s.Message = msg })
	}
}

func exitCodeIs(err error, code int) bool {
	var ee *exec.ExitError
	return errors.As(err, &ee) && ee.ExitCode() == code
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

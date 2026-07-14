package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// ampUpdateControl is a containerised AMP control plane with API credentials,
// the configuration required to reach the instance ADS API for an update. The
// post-update recovery hook is stubbed to a no-op so ExecCommand("update") tests
// don't spawn the real background watcher (which would poll Core/GetStatus).
func ampUpdateControl() *ampControl {
	return &ampControl{
		useContainer:       true,
		container:          "AMP_X",
		ampUser:            "amp",
		containerRuntime:   "docker",
		instance:           "Dune01",
		apiUser:            "admin",
		apiPass:            "pw",
		afterUpdateRestart: func(*ampAPIClient, Executor) {},
	}
}

// TestAmpExecCommand_UpdateCallsUpdateApplication verifies that "update" under
// AMP logs in once and POSTs Core/UpdateApplication to the instance's loopback
// ADS API, wrapped for in-container exec — the same SteamCMD update the AMP
// dashboard "Update" button triggers.
func TestAmpExecCommand_UpdateCallsUpdateApplication(t *testing.T) {
	t.Parallel()
	var updateCmd string
	logins := 0
	exec := &fnExecutor{fn: func(cmd string) (string, error) {
		switch {
		case strings.Contains(cmd, "Core/Login"):
			logins++
			return `{"success":true,"sessionID":"sess"}`, nil
		case strings.Contains(cmd, "Core/UpdateApplication"):
			updateCmd = cmd
			return `{"Id":"task-1","Name":"Updating Application"}`, nil
		default:
			t.Fatalf("unexpected AMP API endpoint in cmd: %q", cmd)
			return "", nil
		}
	}}

	out, err := ampUpdateControl().ExecCommand(context.Background(), exec, "update")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if logins != 1 {
		t.Errorf("logins = %d, want 1", logins)
	}
	if !strings.Contains(updateCmd, "http://127.0.0.1:8081/API/Core/UpdateApplication") {
		t.Errorf("update must hit Core/UpdateApplication on the loopback ADS API, got: %q", updateCmd)
	}
	if !strings.Contains(updateCmd, "docker exec AMP_X") {
		t.Errorf("update API call must be wrapped for in-container exec, got: %q", updateCmd)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("update should return a human-readable confirmation for the UI")
	}
}

// TestAmpExecCommand_UpdateKicksRecovery verifies a successful update triggers
// the post-update recovery hook (which, in production, waits for the update to
// finish then restarts the container) with the authenticated API client.
func TestAmpExecCommand_UpdateKicksRecovery(t *testing.T) {
	t.Parallel()
	exec := &fnExecutor{fn: func(cmd string) (string, error) {
		if strings.Contains(cmd, "Core/Login") {
			return `{"success":true,"sessionID":"sess"}`, nil
		}
		return `{"Id":"task-1"}`, nil
	}}
	kicked := false
	c := ampUpdateControl()
	c.afterUpdateRestart = func(client *ampAPIClient, _ Executor) {
		kicked = true
		if client == nil {
			t.Error("recovery hook must receive the authenticated API client")
		}
	}
	if _, err := c.ExecCommand(context.Background(), exec, "update"); err != nil {
		t.Fatalf("update: %v", err)
	}
	if !kicked {
		t.Error("successful update must kick the post-update recovery hook")
	}
}

// TestAmpExecCommand_UpdateFailureSkipsRecovery verifies a failed update does NOT
// trigger the recovery restart — nothing to recover, and restarting would be a
// surprise side effect of a no-op update.
func TestAmpExecCommand_UpdateFailureSkipsRecovery(t *testing.T) {
	t.Parallel()
	exec := &fnExecutor{fn: func(cmd string) (string, error) {
		if strings.Contains(cmd, "Core/Login") {
			return `{"success":true,"sessionID":"sess"}`, nil
		}
		return `{"Status":false,"Reason":"boom"}`, nil
	}}
	kicked := false
	c := ampUpdateControl()
	c.afterUpdateRestart = func(*ampAPIClient, Executor) { kicked = true }
	if _, err := c.ExecCommand(context.Background(), exec, "update"); err == nil {
		t.Fatal("expected update failure")
	}
	if kicked {
		t.Error("a failed update must not kick the recovery restart")
	}
}

// ── waitForUpdateThenRestart orchestration ──────────────────────────────────

// stepClock returns a nowFn that advances by step on every call, so poll loops
// reach their deadlines deterministically without real sleeps.
func stepClock(step time.Duration) func() time.Time {
	base := time.Unix(1_700_000_000, 0)
	n := 0
	return func() time.Time {
		t := base.Add(time.Duration(n) * step)
		n++
		return t
	}
}

// TestWaitForUpdateThenRestart_WaitsForTaskToClear verifies the watcher restarts
// only after the AMP update task has appeared and then cleared.
func TestWaitForUpdateThenRestart_WaitsForTaskToClear(t *testing.T) {
	t.Parallel()
	// Task present for the first 3 polls, then gone.
	counts := []int{1, 1, 1, 0}
	i := 0
	statusFn := func() (int, error) {
		n := counts[i]
		if i < len(counts)-1 {
			i++
		}
		return n, nil
	}
	restarted := false
	err := waitForUpdateThenRestart(statusFn, func() error { restarted = true; return nil },
		func(time.Duration) {}, stepClock(ampUpdatePollInterval))
	if err != nil {
		t.Fatalf("waitForUpdateThenRestart: %v", err)
	}
	if !restarted {
		t.Error("must restart after the update task clears")
	}
	if i < 3 {
		t.Errorf("restarted too early: only polled %d times before the task cleared", i)
	}
}

// TestWaitForUpdateThenRestart_RestartsIfTaskNeverAppears verifies that if no
// update task ever appears within the grace window, the watcher still restarts
// (so a fast/no-op update isn't left un-recovered forever).
func TestWaitForUpdateThenRestart_RestartsIfTaskNeverAppears(t *testing.T) {
	t.Parallel()
	restarted := false
	err := waitForUpdateThenRestart(
		func() (int, error) { return 0, nil },
		func() error { restarted = true; return nil },
		func(time.Duration) {},
		stepClock(ampUpdateAppearGrace), // each poll jumps a full grace window
	)
	if err != nil {
		t.Fatalf("waitForUpdateThenRestart: %v", err)
	}
	if !restarted {
		t.Error("must restart even when no update task ever appears")
	}
}

// TestWaitForUpdateThenRestart_RestartsAtMaxWait verifies the safety cap: a task
// that never clears still triggers a restart once the max wait elapses.
func TestWaitForUpdateThenRestart_RestartsAtMaxWait(t *testing.T) {
	t.Parallel()
	restarted := false
	err := waitForUpdateThenRestart(
		func() (int, error) { return 1, nil }, // task never clears
		func() error { restarted = true; return nil },
		func(time.Duration) {},
		stepClock(ampUpdateMaxWait), // blow past the cap on the second poll
	)
	if err != nil {
		t.Fatalf("waitForUpdateThenRestart: %v", err)
	}
	if !restarted {
		t.Error("must restart at the max-wait cap even if the task never clears")
	}
}

// TestWaitForUpdateThenRestart_PropagatesRestartError verifies a restart failure
// is surfaced (the caller logs it).
func TestWaitForUpdateThenRestart_PropagatesRestartError(t *testing.T) {
	t.Parallel()
	err := waitForUpdateThenRestart(
		func() (int, error) { return 0, nil },
		func() error { return errors.New("restart boom") },
		func(time.Duration) {},
		stepClock(ampUpdateAppearGrace),
	)
	if err == nil || !strings.Contains(err.Error(), "restart boom") {
		t.Fatalf("expected restart error to propagate, got: %v", err)
	}
}

// TestAMPAPIRunningTaskCount_ParsesAndRetries verifies the running-task count is
// read from Core/GetStatus.RunningTasks and that a session expiry re-logs in.
func TestAMPAPIRunningTaskCount_ParsesAndRetries(t *testing.T) {
	t.Parallel()
	// Count from a populated RunningTasks array.
	exec := &fnExecutor{fn: func(cmd string) (string, error) {
		if strings.Contains(cmd, "Core/Login") {
			return `{"success":true,"sessionID":"s"}`, nil
		}
		return `{"State":75,"RunningTasks":[{"Id":"a"},{"Id":"b"}]}`, nil
	}}
	c := newAMPAPIClient(exec, identityWrap, "u", "p", 0)
	n, err := c.runningTaskCount()
	if err != nil {
		t.Fatalf("runningTaskCount: %v", err)
	}
	if n != 2 {
		t.Errorf("count = %d, want 2", n)
	}

	// Empty array → 0 (update finished).
	execDone := &fnExecutor{fn: func(cmd string) (string, error) {
		if strings.Contains(cmd, "Core/Login") {
			return `{"success":true,"sessionID":"s"}`, nil
		}
		return `{"State":999,"RunningTasks":[]}`, nil
	}}
	cDone := newAMPAPIClient(execDone, identityWrap, "u", "p", 0)
	if n, err := cDone.runningTaskCount(); err != nil || n != 0 {
		t.Fatalf("empty RunningTasks: n=%d err=%v, want 0/nil", n, err)
	}
}

// TestAmpExecCommand_UpdateMissingCredentials verifies update fails fast with a
// clear message — and contacts nothing — when the AMP API creds are unset.
func TestAmpExecCommand_UpdateMissingCredentials(t *testing.T) {
	t.Parallel()
	called := false
	exec := &fnExecutor{fn: func(string) (string, error) { called = true; return "", nil }}
	c := &ampControl{instance: "Dune01", useContainer: true, container: "AMP_X", ampUser: "amp", containerRuntime: "docker"} // no api creds
	_, err := c.ExecCommand(context.Background(), exec, "update")
	if err == nil {
		t.Fatal("expected error when AMP API credentials are not configured")
	}
	if !strings.Contains(err.Error(), "amp_api_user") {
		t.Errorf("error should name the missing config, got: %v", err)
	}
	if called {
		t.Error("must not contact the AMP API without credentials")
	}
}

// TestAmpExecCommand_UpdateRejectionPropagates verifies an AMP ActionResult
// rejection (e.g. already up to date) surfaces its reason rather than being
// swallowed.
func TestAmpExecCommand_UpdateRejectionPropagates(t *testing.T) {
	t.Parallel()
	exec := &fnExecutor{fn: func(cmd string) (string, error) {
		if strings.Contains(cmd, "Core/Login") {
			return `{"success":true,"sessionID":"sess"}`, nil
		}
		return `{"Status":false,"Reason":"Application is already up to date."}`, nil
	}}
	_, err := ampUpdateControl().ExecCommand(context.Background(), exec, "update")
	if err == nil {
		t.Fatal("expected an AMP rejection to propagate")
	}
	if !strings.Contains(err.Error(), "already up to date") {
		t.Errorf("error should surface the AMP reason, got: %v", err)
	}
}

// TestAmpExecCommand_UpdateLoginFailureAborts verifies a login failure aborts
// the update without attempting Core/UpdateApplication.
func TestAmpExecCommand_UpdateLoginFailureAborts(t *testing.T) {
	t.Parallel()
	updateCalled := false
	exec := &fnExecutor{fn: func(cmd string) (string, error) {
		if strings.Contains(cmd, "Core/UpdateApplication") {
			updateCalled = true
		}
		return `{"success":false,"resultReason":"bad creds"}`, nil
	}}
	if _, err := ampUpdateControl().ExecCommand(context.Background(), exec, "update"); err == nil {
		t.Fatal("expected login failure to abort the update")
	}
	if updateCalled {
		t.Error("must not call UpdateApplication when login fails")
	}
}

// TestAmpExecCommand_UnknownCommandStillErrors guards that adding "update" did
// not weaken rejection of genuinely unsupported commands.
func TestAmpExecCommand_UnknownCommandStillErrors(t *testing.T) {
	t.Parallel()
	exec := &fnExecutor{fn: func(string) (string, error) { return "", nil }}
	_, err := ampUpdateControl().ExecCommand(context.Background(), exec, "frobnicate")
	if err == nil || !strings.Contains(err.Error(), "does not support") {
		t.Fatalf("unknown command should be rejected, got: %v", err)
	}
}

// TestAMPAPIUpdateApplication_AcceptsRunningTaskAndVoid verifies the API client
// treats a RunningTask object, a bare {}, and an empty body (void return on some
// AMP builds) all as success — only an explicit ActionResult failure is an error.
func TestAMPAPIUpdateApplication_AcceptsRunningTaskAndVoid(t *testing.T) {
	t.Parallel()
	for _, resp := range []string{`{"Id":"t","Name":"Updating"}`, ``, `{}`} {
		resp := resp
		exec := &fnExecutor{fn: func(cmd string) (string, error) {
			if strings.Contains(cmd, "Core/Login") {
				return `{"success":true,"sessionID":"s"}`, nil
			}
			return resp, nil
		}}
		c := newAMPAPIClient(exec, identityWrap, "u", "p", 0)
		if _, err := c.updateApplication(); err != nil {
			t.Errorf("resp %q: unexpected error: %v", resp, err)
		}
	}
}

// TestAMPAPIUpdateApplication_RetriesOnSessionExpiry verifies a stale-session
// rejection triggers one re-login and a successful retry.
func TestAMPAPIUpdateApplication_RetriesOnSessionExpiry(t *testing.T) {
	t.Parallel()
	logins, updates := 0, 0
	exec := &fnExecutor{fn: func(cmd string) (string, error) {
		if strings.Contains(cmd, "Core/Login") {
			logins++
			return `{"success":true,"sessionID":"s"}`, nil
		}
		updates++
		if updates == 1 {
			return `{"Status":false,"Reason":"Invalid session ID."}`, nil
		}
		return `{"Id":"t","Name":"Updating"}`, nil
	}}
	c := newAMPAPIClient(exec, identityWrap, "u", "p", 0)
	if _, err := c.updateApplication(); err != nil {
		t.Fatalf("updateApplication: %v", err)
	}
	if logins != 2 {
		t.Errorf("logins = %d, want 2 (one re-login on session expiry)", logins)
	}
	if updates != 2 {
		t.Errorf("update attempts = %d, want 2", updates)
	}
}

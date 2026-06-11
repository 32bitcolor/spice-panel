package main

import (
	"errors"
	"testing"
)

// TestRestartProcess_ReExecSucceeds verifies that when the re-exec succeeds,
// the fallback is NOT invoked.
func TestRestartProcess_ReExecSucceeds(t *testing.T) {
	t.Parallel()

	fallbackCalled := false
	restartProcess(
		func() error { return nil },
		func() error { fallbackCalled = true; return nil },
	)
	if fallbackCalled {
		t.Error("fallback was called despite successful re-exec")
	}
}

// TestRestartProcess_FallsBackToRestart verifies that when re-exec fails (e.g.
// on Windows, where syscall.Exec is unsupported), the fallback restart fn is called.
func TestRestartProcess_FallsBackToRestart(t *testing.T) {
	t.Parallel()

	fallbackCalled := false
	restartProcess(
		func() error { return errors.New("exec unsupported") },
		func() error { fallbackCalled = true; return nil },
	)
	if !fallbackCalled {
		t.Error("fallback was not called after re-exec failure")
	}
}

// TestSpawnAndExit_SpawnError verifies that when spawning fails, an error is
// returned and exitFn is NOT called.
func TestSpawnAndExit_SpawnError(t *testing.T) {
	t.Parallel()

	exited := false
	err := spawnAndExit(
		"/nonexistent-bin",
		func(_ string) error { return errors.New("spawn failed") },
		func(_ int) { exited = true },
	)
	if err == nil {
		t.Fatal("expected error when spawn fails")
	}
	if exited {
		t.Error("exitFn should not be called when spawn fails")
	}
}

// TestSpawnAndExit_CallsExit verifies that on a successful spawn, exitFn is
// called with code 0 to terminate the current process.
func TestSpawnAndExit_CallsExit(t *testing.T) {
	t.Parallel()

	exitCode := -1
	err := spawnAndExit(
		"dummy-bin",
		func(_ string) error { return nil },
		func(code int) { exitCode = code },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("exitFn called with code %d, want 0", exitCode)
	}
}

package main

import (
	"testing"
)

func TestServerTimezone_ServerWins(t *testing.T) {

	db := openTestCalibStore(t)
	seedServer(t, db, 1)

	orig := globalStore
	globalStore = db
	defer func() { globalStore = orig }()

	// Set timezone on the server row directly.
	if _, err := db.Exec(`UPDATE servers SET timezone='America/Chicago' WHERE id=1`); err != nil {
		t.Fatalf("set server timezone: %v", err)
	}

	// Seed a restart schedule with a different timezone — server must win.
	if err := saveScheduledRestartConfig(1, scheduledRestartConfig{Timezone: "Europe/London"}); err != nil {
		t.Fatalf("save restart schedule: %v", err)
	}

	if got := serverTimezone(1); got != "America/Chicago" {
		t.Errorf("expected America/Chicago, got %q", got)
	}
}

func TestServerTimezone_FallsBackToSchedule(t *testing.T) {

	db := openTestCalibStore(t)
	seedServer(t, db, 1)

	orig := globalStore
	globalStore = db
	defer func() { globalStore = orig }()

	// Server timezone is blank — should fall back to schedule timezone.
	if err := saveScheduledRestartConfig(1, scheduledRestartConfig{Timezone: "Asia/Tokyo"}); err != nil {
		t.Fatalf("save restart schedule: %v", err)
	}

	if got := serverTimezone(1); got != "Asia/Tokyo" {
		t.Errorf("expected Asia/Tokyo (from schedule fallback), got %q", got)
	}
}

func TestServerTimezone_BothEmpty(t *testing.T) {

	db := openTestCalibStore(t)
	seedServer(t, db, 1)

	orig := globalStore
	globalStore = db
	defer func() { globalStore = orig }()

	if got := serverTimezone(1); got != "" {
		t.Errorf("expected empty string when both unset, got %q", got)
	}
}

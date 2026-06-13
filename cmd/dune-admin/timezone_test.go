package main

import (
	"testing"
	"time"
)

// TestTimezoneDatabaseAvailable guards the `_ "time/tzdata"` import in main.go.
// Scheduled restart validation (handlers_scheduled_restart.go) calls
// time.LoadLocation on the user-supplied IANA zone; on a minimal container
// without the OS tzdata package this fails for valid zones like "Europe/London"
// (#204). The embedded database makes these loads succeed regardless of host.
func TestTimezoneDatabaseAvailable(t *testing.T) {
	zones := []string{
		"Europe/London",
		"America/New_York",
		"Asia/Tokyo",
		"Australia/Sydney",
		"UTC",
	}
	for _, zone := range zones {
		if _, err := time.LoadLocation(zone); err != nil {
			t.Errorf("LoadLocation(%q) failed — is the time/tzdata import present? %v", zone, err)
		}
	}
}

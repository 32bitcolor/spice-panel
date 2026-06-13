package main

import (
	"testing"
	"time"
)

// TestAgeSecondsFromStartTime covers the best-effort uptime parse used to fill
// ServerRow.AgeSeconds on kubectl (#203). Unparseable/absent/future timestamps
// yield 0 so the UI shows "—" exactly as before — no regression.
func TestAgeSecondsFromStartTime(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		ts   string
		want int
	}{
		{"empty", "", 0},
		{"garbage", "not-a-time", 0},
		{"one hour ago", "2026-06-13T11:00:00Z", 3600},
		{"future clamps to zero", "2026-06-13T13:00:00Z", 0},
		{"with offset", "2026-06-13T11:00:00+00:00", 3600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ageSecondsFromStartTime(tt.ts, now); got != tt.want {
				t.Errorf("ageSecondsFromStartTime(%q) = %d, want %d", tt.ts, got, tt.want)
			}
		})
	}
}

package main

import "testing"

// TestIntelGrantDelta verifies the headroom clamp used by cmdAwardIntelCtx so a
// battlepass grant can never push a character above maxIntelPoints, leaving
// unspendable intel (#208). The delta is how much intel is actually added.
func TestIntelGrantDelta(t *testing.T) {
	tests := []struct {
		name      string
		current   int64
		requested int64
		want      int64
	}{
		{"under cap, fits", 100, 50, 50},
		{"under cap, would exceed clamps to headroom", maxIntelPoints - 10, 100, 10},
		{"exactly at cap grants nothing", maxIntelPoints, 100, 0},
		{"over cap (defensive) grants nothing", maxIntelPoints + 500, 100, 0},
		{"zero request", 100, 0, 0},
		{"negative request never reduces", 100, -50, 0},
		{"empty character fills to cap", 0, maxIntelPoints + 1000, maxIntelPoints},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intelGrantDelta(tt.current, tt.requested); got != tt.want {
				t.Errorf("intelGrantDelta(%d, %d) = %d, want %d", tt.current, tt.requested, got, tt.want)
			}
		})
	}
}

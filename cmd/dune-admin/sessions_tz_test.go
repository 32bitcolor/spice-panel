package main

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func mustExecSession(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("mustExecSession: %v", err)
	}
}

// TestGetActivityTrendCounts_LocalTimezone verifies that sessions straddling a
// UTC midnight are bucketed into the correct LOCAL day when a non-UTC timezone
// is supplied (issue #236).
func TestGetActivityTrendCounts_LocalTimezone(t *testing.T) {
	t.Parallel()
	db := openTestSessionDB(t)
	ctx := context.Background()

	// Session A: 2026-01-01T23:30:00Z → UTC Jan 1, America/New_York Jan 1 18:30
	// Session B: 2026-01-02T01:30:00Z → UTC Jan 2, America/New_York Jan 1 20:30
	// With UTC bucketing these land on two different UTC days.
	// With NY bucketing both must land on local 2026-01-01.
	mustExecSession(t, db, `INSERT INTO play_sessions(server_id, account_id, started_at) VALUES(1,1,'2026-01-01T23:30:00Z')`)
	mustExecSession(t, db, `INSERT INTO play_sessions(server_id, account_id, started_at) VALUES(1,2,'2026-01-02T01:30:00Z')`)

	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load NY tz: %v", err)
	}
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, nyc)
	counts, err := getActivityTrendCounts(ctx, db, 1, since, nyc)
	if err != nil {
		t.Fatalf("getActivityTrendCounts: %v", err)
	}
	if counts["2026-01-01"] != 2 {
		t.Errorf("expected 2 sessions on local 2026-01-01, got %+v", counts)
	}
	if counts["2026-01-02"] != 0 {
		t.Errorf("expected 0 sessions on local 2026-01-02 (both shifted to Jan 1), got %+v", counts)
	}
}

// TestGetActivityTrendCounts_UTCPreserved verifies that the UTC (zero-offset)
// fallback still buckets sessions on their UTC date.
func TestGetActivityTrendCounts_UTCPreserved(t *testing.T) {
	t.Parallel()
	db := openTestSessionDB(t)
	ctx := context.Background()

	mustExecSession(t, db, `INSERT INTO play_sessions(server_id, account_id, started_at) VALUES(1,1,'2026-01-01T23:30:00Z')`)
	mustExecSession(t, db, `INSERT INTO play_sessions(server_id, account_id, started_at) VALUES(1,2,'2026-01-02T01:30:00Z')`)

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	counts, err := getActivityTrendCounts(ctx, db, 1, since, time.UTC)
	if err != nil {
		t.Fatalf("getActivityTrendCounts UTC: %v", err)
	}
	if counts["2026-01-01"] != 1 {
		t.Errorf("expected 1 UTC session on 2026-01-01, got %+v", counts)
	}
	if counts["2026-01-02"] != 1 {
		t.Errorf("expected 1 UTC session on 2026-01-02, got %+v", counts)
	}
}

// TestSessionSummary_LocalTrend verifies that sessionSummaryWithLoc anchors the
// trend window on local-timezone "today" rather than UTC today (issue #236).
func TestSessionSummary_LocalTrend(t *testing.T) {
	t.Parallel()
	db := openTestSessionDB(t)
	ctx := context.Background()

	// Session starts at UTC Jan 2 02:00 = NY Jan 1 21:00.
	mustExecSession(t, db, `INSERT INTO play_sessions(server_id, account_id, started_at) VALUES(1,1,'2026-01-02T02:00:00Z')`)

	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load NY tz: %v", err)
	}
	// Fake "now" = NY Jan 1 23:00 so "today" locally is Jan 1.
	now := time.Date(2026, 1, 1, 23, 0, 0, 0, nyc)
	_, trend := sessionSummaryWithLoc(ctx, db, 1, 7, now, nyc)

	var jan1Count int64
	for _, p := range trend {
		if p.Day == "2026-01-01" {
			jan1Count = p.Count
		}
	}
	if jan1Count != 1 {
		t.Errorf("expected local Jan 1 to have 1 session; trend=%+v", trend)
	}
}

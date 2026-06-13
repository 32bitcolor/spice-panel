package main

import (
	"testing"
	"time"
)

func testEventStore(t *testing.T) *eventStore {
	t.Helper()
	s, err := openEventStore(":memory:")
	if err != nil {
		t.Fatalf("openEventStore: %v", err)
	}
	t.Cleanup(func() { _ = s.db.Close() })
	return s
}

func TestEventStoreCreatePreservesSchedule(t *testing.T) {
	s := testEventStore(t)

	def := eventDefinition{
		Name:          "Race 1",
		Type:          eventTypeZoneRace,
		PollSeconds:   15,
		JitterSeconds: 5,
	}
	created, err := s.create(def)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.PollSeconds != 15 {
		t.Errorf("PollSeconds = %d, want 15", created.PollSeconds)
	}
	if created.JitterSeconds != 5 {
		t.Errorf("JitterSeconds = %d, want 5", created.JitterSeconds)
	}
}

func TestEventStoreUpdatePreservesSchedule(t *testing.T) {
	s := testEventStore(t)

	def := eventDefinition{Name: "M1", Type: eventTypeMilestone, PollSeconds: 10, JitterSeconds: 2}
	created, err := s.create(def)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	created.PollSeconds = 30
	created.JitterSeconds = 10
	updated, err := s.update(*created)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.PollSeconds != 30 {
		t.Errorf("PollSeconds = %d, want 30", updated.PollSeconds)
	}
	if updated.JitterSeconds != 10 {
		t.Errorf("JitterSeconds = %d, want 10", updated.JitterSeconds)
	}
}

func TestEventStoreDefaultSchedule(t *testing.T) {
	s := testEventStore(t)

	// Create without specifying schedule — should get defaults.
	created, err := s.create(eventDefinition{Name: "Z", Type: eventTypeZoneRace})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.PollSeconds != 7 {
		t.Errorf("default PollSeconds = %d, want 7", created.PollSeconds)
	}
	if created.JitterSeconds != 3 {
		t.Errorf("default JitterSeconds = %d, want 3", created.JitterSeconds)
	}
}

func TestEventStoreListIncludesSchedule(t *testing.T) {
	s := testEventStore(t)

	if _, err := s.create(eventDefinition{Name: "A", Type: eventTypeZoneRace, PollSeconds: 20, JitterSeconds: 4}); err != nil {
		t.Fatalf("create: %v", err)
	}

	list, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0].PollSeconds != 20 {
		t.Errorf("PollSeconds = %d, want 20", list[0].PollSeconds)
	}
}

// ── reward-grant retry lifecycle ──────────────────────────────────────────────

// firstClaim fetches the single claim for an event, failing the test otherwise.
func firstClaim(t *testing.T, s *eventStore, eventID int64) eventClaimRecord {
	t.Helper()
	claims, err := s.listClaims(eventID)
	if err != nil {
		t.Fatalf("listClaims: %v", err)
	}
	if len(claims) != 1 {
		t.Fatalf("want 1 claim, got %d", len(claims))
	}
	return claims[0]
}

func TestRecordFailed_FirstFailureIsPendingWithBackoff(t *testing.T) {
	s := testEventStore(t)
	if err := s.recordFailed(1, 1, 101, "inventory full"); err != nil {
		t.Fatalf("recordFailed: %v", err)
	}
	c := firstClaim(t, s, 1)
	if c.Status != eventClaimStatusPending {
		t.Errorf("status = %q, want pending", c.Status)
	}
	if c.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", c.Attempts)
	}
	if c.LastError != "inventory full" {
		t.Errorf("last_error = %q", c.LastError)
	}
	next, err := time.Parse(time.RFC3339, c.NextAttemptAt)
	if err != nil {
		t.Fatalf("parse next_attempt_at %q: %v", c.NextAttemptAt, err)
	}
	wantMin := time.Now().Add(eventGrantRetryBackoff - 2*time.Minute)
	wantMax := time.Now().Add(eventGrantRetryBackoff + 2*time.Minute)
	if next.Before(wantMin) || next.After(wantMax) {
		t.Errorf("next_attempt_at = %v, want ~now+24h", next)
	}
}

func TestRecordFailed_ExhaustsAfterMaxAttempts(t *testing.T) {
	s := testEventStore(t)
	for i := 0; i < eventGrantMaxAttempts; i++ {
		if err := s.recordFailed(1, 1, 101, "still full"); err != nil {
			t.Fatalf("recordFailed #%d: %v", i, err)
		}
	}
	c := firstClaim(t, s, 1)
	if c.Status != eventClaimStatusExhausted {
		t.Errorf("status = %q, want exhausted", c.Status)
	}
	if c.Attempts != eventGrantMaxAttempts {
		t.Errorf("attempts = %d, want %d", c.Attempts, eventGrantMaxAttempts)
	}
	if c.NextAttemptAt != "" {
		t.Errorf("next_attempt_at = %q, want empty for exhausted", c.NextAttemptAt)
	}
}

func TestRecordGranted_ClearsNextAttempt(t *testing.T) {
	s := testEventStore(t)
	if err := s.recordFailed(1, 1, 101, "full"); err != nil {
		t.Fatalf("recordFailed: %v", err)
	}
	if err := s.recordGranted(1, 1, 101); err != nil {
		t.Fatalf("recordGranted: %v", err)
	}
	c := firstClaim(t, s, 1)
	if c.Status != eventClaimStatusGranted {
		t.Errorf("status = %q, want granted", c.Status)
	}
	if c.NextAttemptAt != "" {
		t.Errorf("next_attempt_at = %q, want empty after grant", c.NextAttemptAt)
	}
}

func TestClaimExists_Matrix(t *testing.T) {
	tests := []struct {
		name  string
		setup func(s *eventStore)
		want  bool
	}{
		{"no claim", func(_ *eventStore) {}, false},
		{"granted blocks", func(s *eventStore) { _ = s.recordGranted(1, 1, 101) }, true},
		{"pending does not block", func(s *eventStore) { _ = s.recordFailed(1, 1, 101, "full") }, false},
		{"exhausted blocks", func(s *eventStore) {
			for i := 0; i < eventGrantMaxAttempts; i++ {
				_ = s.recordFailed(1, 1, 101, "full")
			}
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testEventStore(t)
			tt.setup(s)
			got, err := s.claimExists(1, 1, 101)
			if err != nil {
				t.Fatalf("claimExists: %v", err)
			}
			if got != tt.want {
				t.Errorf("claimExists = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListRetryableClaims_FiltersCorrectly(t *testing.T) {
	s := testEventStore(t)

	// Pending and due (next_attempt_at in the past).
	if err := s.recordFailed(1, 1, 101, "full"); err != nil {
		t.Fatalf("recordFailed: %v", err)
	}
	// Granted — never retryable.
	if err := s.recordGranted(2, 1, 102); err != nil {
		t.Fatalf("recordGranted: %v", err)
	}
	// Exhausted — manual-only, never retryable.
	for i := 0; i < eventGrantMaxAttempts; i++ {
		if err := s.recordFailed(3, 1, 103, "full"); err != nil {
			t.Fatalf("recordFailed: %v", err)
		}
	}

	// Before the backoff window: nothing is due yet.
	none, err := s.listRetryableClaims(time.Now())
	if err != nil {
		t.Fatalf("listRetryableClaims: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("want 0 due claims now, got %d", len(none))
	}

	// After the backoff window: only the pending claim is due.
	due, err := s.listRetryableClaims(time.Now().Add(eventGrantRetryBackoff + time.Hour))
	if err != nil {
		t.Fatalf("listRetryableClaims: %v", err)
	}
	if len(due) != 1 {
		t.Fatalf("want 1 due claim, got %d", len(due))
	}
	if due[0].AccountID != 101 || due[0].EventID != 1 {
		t.Errorf("unexpected due claim %+v", due[0])
	}
}

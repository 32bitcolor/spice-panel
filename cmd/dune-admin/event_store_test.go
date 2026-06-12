package main

import (
	"testing"
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

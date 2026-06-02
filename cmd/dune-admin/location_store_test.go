package main

import (
	"math"
	"testing"
)

// openMemLocationStore opens an in-memory SQLite location store for testing.
func openMemLocationStore(t *testing.T) *locationStore {
	t.Helper()
	s, err := openLocationStore(":memory:")
	if err != nil {
		t.Fatalf("openLocationStore: %v", err)
	}
	t.Cleanup(func() { _ = s.close() })
	return s
}

func TestOpenLocationStore_SeedsFromCheatLocationsWhenEmpty(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	locs, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(locs) != len(cheatLocations) {
		t.Fatalf("want %d seeded locations, got %d", len(cheatLocations), len(locs))
	}

	// Verify a known seed entry by name and approximate coordinates.
	var found bool
	for _, l := range locs {
		if l.Name == "Windsack" {
			found = true
			if math.Abs(l.X-974276.75) > 0.01 {
				t.Fatalf("Windsack X: want 974276.75, got %f", l.X)
			}
		}
	}
	if !found {
		t.Fatal("seed entry 'Windsack' not found in list")
	}
}

func TestOpenLocationStore_DoesNotReseedWhenNotEmpty(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	// Add a custom entry, then close and reopen (seeding skipped when non-empty).
	if err := s.upsert("Custom", 1, 2, 3); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	locs, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// Should have seeds + our custom entry.
	if len(locs) != len(cheatLocations)+1 {
		t.Fatalf("want %d locations, got %d", len(cheatLocations)+1, len(locs))
	}
}

func TestLocationStore_List_ReturnsNonNilOnEmpty(t *testing.T) {
	t.Parallel()
	// Use a store with seeds cleared so we can test the empty-slice guarantee.
	s := openMemLocationStore(t)
	for _, l := range cheatLocations {
		if err := s.delete(l.Name); err != nil {
			t.Fatalf("delete: %v", err)
		}
	}
	locs, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if locs == nil {
		t.Fatal("list() must return non-nil empty slice, not nil")
	}
	if len(locs) != 0 {
		t.Fatalf("want 0 locations after delete-all, got %d", len(locs))
	}
}

func TestLocationStore_Upsert_AddsNewLocation(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	if err := s.upsert("TestSpot", 100.5, 200.75, 300.0); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	locs, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var found *teleportLocation
	for i := range locs {
		if locs[i].Name == "TestSpot" {
			found = &locs[i]
			break
		}
	}
	if found == nil {
		t.Fatal("upserted location 'TestSpot' not found")
	}
	if math.Abs(found.X-100.5) > 0.001 || math.Abs(found.Y-200.75) > 0.001 || math.Abs(found.Z-300.0) > 0.001 {
		t.Fatalf("coordinates wrong: got %+v", found)
	}
}

func TestLocationStore_Upsert_UpdatesExistingByName(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	if err := s.upsert("Windsack", 1, 2, 3); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	locs, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var found *teleportLocation
	for i := range locs {
		if locs[i].Name == "Windsack" {
			found = &locs[i]
			break
		}
	}
	if found == nil {
		t.Fatal("Windsack not found after update")
	}
	if math.Abs(found.X-1) > 0.001 || math.Abs(found.Y-2) > 0.001 || math.Abs(found.Z-3) > 0.001 {
		t.Fatalf("updated coordinates wrong: got %+v", found)
	}
	// Total count should be unchanged (upsert, not insert).
	if len(locs) != len(cheatLocations) {
		t.Fatalf("count changed after upsert: want %d, got %d", len(cheatLocations), len(locs))
	}
}

func TestLocationStore_Rename_UpdatesName(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	if err := s.rename("Windsack", "WindsackRenamed"); err != nil {
		t.Fatalf("rename: %v", err)
	}

	locs, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var oldFound, newFound bool
	for _, l := range locs {
		if l.Name == "Windsack" {
			oldFound = true
		}
		if l.Name == "WindsackRenamed" {
			newFound = true
		}
	}
	if oldFound {
		t.Fatal("old name 'Windsack' still present after rename")
	}
	if !newFound {
		t.Fatal("new name 'WindsackRenamed' not found after rename")
	}
}

func TestLocationStore_Rename_UnknownNameReturnsError(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	if err := s.rename("NoSuchPlace", "Elsewhere"); err == nil {
		t.Fatal("expected error renaming unknown location, got nil")
	}
}

func TestLocationStore_Delete_RemovesLocation(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	if err := s.delete("Windsack"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	locs, err := s.list()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, l := range locs {
		if l.Name == "Windsack" {
			t.Fatal("deleted location 'Windsack' still present")
		}
	}
	if len(locs) != len(cheatLocations)-1 {
		t.Fatalf("want %d locations after delete, got %d", len(cheatLocations)-1, len(locs))
	}
}

func TestLocationStore_Delete_UnknownNameReturnsError(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	if err := s.delete("NoSuchPlace"); err == nil {
		t.Fatal("expected error deleting unknown location, got nil")
	}
}

func TestLocationStore_Upsert_EmptyNameReturnsError(t *testing.T) {
	t.Parallel()
	s := openMemLocationStore(t)

	if err := s.upsert("", 1, 2, 3); err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

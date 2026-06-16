package main

import (
	"database/sql"
	"testing"
)

func openTestCalibStore(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openUnifiedStore(":memory:")
	if err != nil {
		t.Fatalf("openUnifiedStore: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedServer(t *testing.T, db *sql.DB, ids ...int) {
	t.Helper()
	for _, id := range ids {
		if _, err := db.Exec(`INSERT INTO servers(id, name) VALUES(?, 'test')`, id); err != nil {
			t.Fatalf("seed server %d: %v", id, err)
		}
	}
}

func TestMapCalibration_RoundTrip(t *testing.T) {
	t.Parallel()
	db := openTestCalibStore(t)
	seedServer(t, db, 1)

	in := mapCalibration{
		MapKey: "HaggaBasin",
		MinX:   -437871, MaxX: 350539,
		MinY: -462011, MaxY: 376267,
		FlipY: true,
	}
	if err := saveMapCalibration(db, 1, in); err != nil {
		t.Fatalf("saveMapCalibration: %v", err)
	}

	got, ok, err := loadMapCalibration(db, 1, "HaggaBasin")
	if err != nil {
		t.Fatalf("loadMapCalibration: %v", err)
	}
	if !ok {
		t.Fatal("expected calibration to exist, got ok=false")
	}
	if got.MinX != in.MinX || got.MaxX != in.MaxX || got.MinY != in.MinY || got.MaxY != in.MaxY {
		t.Errorf("bounds mismatch: got %+v, want %+v", got, in)
	}
	if got.FlipY != in.FlipY {
		t.Errorf("flipY mismatch: got %v, want %v", got.FlipY, in.FlipY)
	}
}

func TestMapCalibration_NotFound(t *testing.T) {
	t.Parallel()
	db := openTestCalibStore(t)
	seedServer(t, db, 1)

	_, ok, err := loadMapCalibration(db, 1, "HaggaBasin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected ok=false for missing calibration")
	}
}

func TestMapCalibration_Upsert(t *testing.T) {
	t.Parallel()
	db := openTestCalibStore(t)
	seedServer(t, db, 1)

	first := mapCalibration{MapKey: "HaggaBasin", MinX: -100, MaxX: 100, MinY: -100, MaxY: 100}
	if err := saveMapCalibration(db, 1, first); err != nil {
		t.Fatalf("first save: %v", err)
	}
	second := mapCalibration{MapKey: "HaggaBasin", MinX: -200, MaxX: 200, MinY: -200, MaxY: 200, FlipY: true}
	if err := saveMapCalibration(db, 1, second); err != nil {
		t.Fatalf("second save: %v", err)
	}

	got, ok, err := loadMapCalibration(db, 1, "HaggaBasin")
	if err != nil || !ok {
		t.Fatalf("load after upsert: ok=%v err=%v", ok, err)
	}
	if got.MinX != -200 {
		t.Errorf("expected upserted MinX=-200, got %v", got.MinX)
	}
}

func TestMapCalibration_ServerIsolation(t *testing.T) {
	t.Parallel()
	db := openTestCalibStore(t)
	seedServer(t, db, 1, 2)

	srv1 := mapCalibration{MapKey: "HaggaBasin", MinX: -100, MaxX: 100, MinY: -100, MaxY: 100}
	if err := saveMapCalibration(db, 1, srv1); err != nil {
		t.Fatalf("save srv1: %v", err)
	}

	_, ok, err := loadMapCalibration(db, 2, "HaggaBasin")
	if err != nil {
		t.Fatalf("load srv2: %v", err)
	}
	if ok {
		t.Error("server 2 should not see server 1's calibration")
	}
}

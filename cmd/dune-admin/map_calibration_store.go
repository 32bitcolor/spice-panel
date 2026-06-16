package main

import (
	"database/sql"
	"fmt"
)

// mapCalibration holds per-map world-coordinate bounds used by the Live Map
// to project Unreal cm coordinates onto the map image.
type mapCalibration struct {
	MapKey string  `json:"map_key"`
	MinX   float64 `json:"min_x"`
	MaxX   float64 `json:"max_x"`
	MinY   float64 `json:"min_y"`
	MaxY   float64 `json:"max_y"`
	FlipX  bool    `json:"flip_x"`
	FlipY  bool    `json:"flip_y"`
}

func initMapCalibrationSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS map_calibration (
			map_key   TEXT    NOT NULL,
			server_id INTEGER NOT NULL DEFAULT 0 REFERENCES servers(id) ON DELETE CASCADE,
			min_x     REAL    NOT NULL DEFAULT 0,
			max_x     REAL    NOT NULL DEFAULT 0,
			min_y     REAL    NOT NULL DEFAULT 0,
			max_y     REAL    NOT NULL DEFAULT 0,
			flip_x    INTEGER NOT NULL DEFAULT 0,
			flip_y    INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (map_key, server_id)
		)`)
	return err
}

// loadMapCalibration returns the stored calibration for (serverID, mapKey).
// ok=false when no calibration has been saved yet.
func loadMapCalibration(db *sql.DB, serverID int, mapKey string) (mapCalibration, bool, error) {
	var c mapCalibration
	var flipX, flipY int
	err := db.QueryRow(`
		SELECT map_key, min_x, max_x, min_y, max_y, flip_x, flip_y
		FROM map_calibration
		WHERE server_id = ? AND map_key = ?`, serverID, mapKey).
		Scan(&c.MapKey, &c.MinX, &c.MaxX, &c.MinY, &c.MaxY, &flipX, &flipY)
	if err == sql.ErrNoRows {
		return mapCalibration{}, false, nil
	}
	if err != nil {
		return mapCalibration{}, false, fmt.Errorf("load map calibration: %w", err)
	}
	c.FlipX = flipX != 0
	c.FlipY = flipY != 0
	return c, true, nil
}

// saveMapCalibration upserts the calibration for (serverID, c.MapKey).
func saveMapCalibration(db *sql.DB, serverID int, c mapCalibration) error {
	flipX, flipY := 0, 0
	if c.FlipX {
		flipX = 1
	}
	if c.FlipY {
		flipY = 1
	}
	_, err := db.Exec(`
		INSERT INTO map_calibration(map_key, server_id, min_x, max_x, min_y, max_y, flip_x, flip_y)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(map_key, server_id) DO UPDATE SET
			min_x = excluded.min_x, max_x = excluded.max_x,
			min_y = excluded.min_y, max_y = excluded.max_y,
			flip_x = excluded.flip_x, flip_y = excluded.flip_y`,
		c.MapKey, serverID, c.MinX, c.MaxX, c.MinY, c.MaxY, flipX, flipY)
	if err != nil {
		return fmt.Errorf("save map calibration: %w", err)
	}
	return nil
}

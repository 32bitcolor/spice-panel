package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// give_packs_columns.go stores the give-items pack library as two typed child
// tables (give_packs + give_pack_items) instead of the give_packs_config.packs_json
// blob. Both tables are server-scoped and preserve slice order via a position
// column. The givePack struct and its json tags are unchanged; only storage moves
// to columns. give_packs_config.packs_json is kept (written as '[]') but no longer
// authoritative once migrated (see migrateGivePacksColumns).

const givePacksColumnsSchema = `
CREATE TABLE IF NOT EXISTS give_packs (
	server_id TEXT NOT NULL, pack_id TEXT NOT NULL,
	name TEXT NOT NULL DEFAULT '', category TEXT NOT NULL DEFAULT '',
	tier INTEGER NOT NULL DEFAULT 0, position INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (server_id, pack_id)
);
CREATE TABLE IF NOT EXISTS give_pack_items (
	server_id TEXT NOT NULL, pack_id TEXT NOT NULL, position INTEGER NOT NULL DEFAULT 0,
	template TEXT NOT NULL DEFAULT '', qty INTEGER NOT NULL DEFAULT 0, quality INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (server_id, pack_id, position)
);`

// initGivePacksColumnsSchema creates the give_packs and give_pack_items tables.
// Idempotent.
func initGivePacksColumnsSchema(db *sql.DB) error {
	if _, err := db.Exec(givePacksColumnsSchema); err != nil {
		return fmt.Errorf("init give-packs columns schema: %w", err)
	}
	return nil
}

// saveGivePacksColumns replaces all packs for serverID with packs, preserving
// slice order via the position column. Existing rows for serverID are deleted
// first so the write is a full replacement (matching the blob's all-or-nothing
// semantics).
func saveGivePacksColumns(db dbExecer, serverID string, packs []givePack) error {
	if _, err := db.Exec(`DELETE FROM give_pack_items WHERE server_id = ?`, serverID); err != nil {
		return fmt.Errorf("clear give_pack_items %s: %w", serverID, err)
	}
	if _, err := db.Exec(`DELETE FROM give_packs WHERE server_id = ?`, serverID); err != nil {
		return fmt.Errorf("clear give_packs %s: %w", serverID, err)
	}
	for pos, pack := range packs {
		if _, err := db.Exec(`INSERT INTO give_packs (server_id, pack_id, name, category, tier, position)
			VALUES (?, ?, ?, ?, ?, ?)`,
			serverID, pack.ID, pack.Name, pack.Category, pack.Tier, pos); err != nil {
			return fmt.Errorf("insert give_pack %s/%s: %w", serverID, pack.ID, err)
		}
		for itemPos, item := range pack.Items {
			if _, err := db.Exec(`INSERT INTO give_pack_items
				(server_id, pack_id, position, template, qty, quality)
				VALUES (?, ?, ?, ?, ?, ?)`,
				serverID, pack.ID, itemPos, item.Template, item.Qty, item.Quality); err != nil {
				return fmt.Errorf("insert give_pack_item %s/%s[%d]: %w", serverID, pack.ID, itemPos, err)
			}
		}
	}
	return nil
}

// loadGivePacksColumns rebuilds the ordered []givePack for serverID from the two
// child tables. Items are fetched once and grouped by pack_id in Go to avoid a
// query-during-rows-iteration conflict on the same connection.
func loadGivePacksColumns(db dbRowQueryer, serverID string) ([]givePack, error) {
	q, ok := db.(interface {
		Query(query string, args ...any) (*sql.Rows, error)
	})
	if !ok {
		return nil, fmt.Errorf("loadGivePacksColumns: db does not support Query")
	}
	packs, order, err := loadGivePackRows(q, serverID)
	if err != nil {
		return nil, err
	}
	if err := attachGivePackItems(q, serverID, packs); err != nil {
		return nil, err
	}
	out := make([]givePack, 0, len(order))
	for _, id := range order {
		out = append(out, *packs[id])
	}
	return out, nil
}

type givePacksQueryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// loadGivePackRows reads give_packs for serverID ordered by position, returning a
// pack_id→pack map (items empty) plus the pack_id order slice.
func loadGivePackRows(db givePacksQueryer, serverID string) (map[string]*givePack, []string, error) {
	rows, err := db.Query(`SELECT pack_id, name, category, tier FROM give_packs
		WHERE server_id = ? ORDER BY position`, serverID)
	if err != nil {
		return nil, nil, fmt.Errorf("query give_packs %s: %w", serverID, err)
	}
	defer func() { _ = rows.Close() }()
	packs := make(map[string]*givePack)
	var order []string
	for rows.Next() {
		var p givePack
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Tier); err != nil {
			return nil, nil, fmt.Errorf("scan give_pack: %w", err)
		}
		p.Items = []welcomePackageItem{}
		packs[p.ID] = &p
		order = append(order, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate give_packs: %w", err)
	}
	return packs, order, nil
}

// attachGivePackItems reads all give_pack_items for serverID ordered by position
// and appends each into its parent pack. Items for unknown packs are skipped.
func attachGivePackItems(db givePacksQueryer, serverID string, packs map[string]*givePack) error {
	rows, err := db.Query(`SELECT pack_id, template, qty, quality FROM give_pack_items
		WHERE server_id = ? ORDER BY pack_id, position`, serverID)
	if err != nil {
		return fmt.Errorf("query give_pack_items %s: %w", serverID, err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var packID string
		var item welcomePackageItem
		if err := rows.Scan(&packID, &item.Template, &item.Qty, &item.Quality); err != nil {
			return fmt.Errorf("scan give_pack_item: %w", err)
		}
		if p, ok := packs[packID]; ok {
			p.Items = append(p.Items, item)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate give_pack_items: %w", err)
	}
	return nil
}

// legacyGivePacksBlob is one decoded give_packs_config.packs_json row.
type legacyGivePacksBlob struct {
	serverID string
	packs    []givePack
}

// readLegacyGivePacksBlobs decodes every give_packs_config.packs_json blob into a
// typed []givePack keyed by server_id. Empty blobs decode to a nil slice. The
// rows are fully buffered so callers can write back without a query-during-rows
// conflict on the same transaction.
func readLegacyGivePacksBlobs(tx *sql.Tx) ([]legacyGivePacksBlob, error) {
	rows, err := tx.Query(`SELECT server_id, packs_json FROM give_packs_config`)
	if err != nil {
		return nil, fmt.Errorf("read legacy give_packs blobs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []legacyGivePacksBlob
	for rows.Next() {
		var rec legacyGivePacksBlob
		var blob string
		if err := rows.Scan(&rec.serverID, &blob); err != nil {
			return nil, fmt.Errorf("scan legacy give_packs: %w", err)
		}
		if blob != "" {
			if err := json.Unmarshal([]byte(blob), &rec.packs); err != nil {
				return nil, fmt.Errorf("unmarshal legacy give_packs %s: %w", rec.serverID, err)
			}
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// migrateGivePacksColumns translates each legacy give_packs_config.packs_json blob
// into the typed give_packs/give_pack_items tables, once, guarded by the
// migrated:give_packs_columns marker. After this runs the blob is never read again.
func migrateGivePacksColumns(db *sql.DB) error {
	return runColumnMigrationOnce(db, "migrated:give_packs_columns", func(tx *sql.Tx) error {
		blobs, err := readLegacyGivePacksBlobs(tx)
		if err != nil {
			return err
		}
		for _, rec := range blobs {
			if err := saveGivePacksColumns(tx, rec.serverID, rec.packs); err != nil {
				return err
			}
		}
		return nil
	})
}

package main

import (
	"testing"
)

// setupGivePacksStore wires a fresh in-memory store into givePacksStoreDB and
// restores nil on cleanup. NOT parallel — mutates a package global.
func setupGivePacksStore(t *testing.T) *givePacksStore {
	t.Helper()
	s := openMemGivePacksStore(t)
	givePacksStoreDB = s
	t.Cleanup(func() { givePacksStoreDB = nil })
	return s
}

// ── validateGivePacks ────────────────────────────────────────────────────────

func TestValidateGivePacks_Valid(t *testing.T) {
	t.Parallel()
	packs := []givePack{
		{ID: "starter-t1", Name: "T1 Starter", Category: "Starter", Tier: 1, Items: []welcomePackageItem{
			{Template: "Ammo", Qty: 100, Quality: 0},
		}},
		{ID: "buggy-t6", Name: "Buggy T6", Category: "Buggy", Tier: 6, Items: []welcomePackageItem{
			{Template: "BuggyBoost_6", Qty: 1, Quality: 0},
		}},
	}
	if err := validateGivePacks(packs); err != nil {
		t.Fatalf("unexpected error for valid packs: %v", err)
	}
}

func TestValidateGivePacks_EmptySliceIsValid(t *testing.T) {
	t.Parallel()
	// An empty pack list is valid — operator deleted everything intentionally.
	if err := validateGivePacks([]givePack{}); err != nil {
		t.Fatalf("empty packs should be valid: %v", err)
	}
}

func TestValidateGivePacks_EmptyID(t *testing.T) {
	t.Parallel()
	packs := []givePack{{ID: "", Name: "No ID", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: 1, Quality: 0}}}}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestValidateGivePacks_DuplicateID(t *testing.T) {
	t.Parallel()
	packs := []givePack{
		{ID: "dup", Name: "A", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: 1, Quality: 0}}},
		{ID: "dup", Name: "B", Category: "Y", Tier: 2, Items: []welcomePackageItem{{Template: "B", Qty: 1, Quality: 0}}},
	}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for duplicate id")
	}
}

func TestValidateGivePacks_EmptyName(t *testing.T) {
	t.Parallel()
	packs := []givePack{{ID: "ok-id", Name: "", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: 1, Quality: 0}}}}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidateGivePacks_EmptyCategory(t *testing.T) {
	t.Parallel()
	packs := []givePack{{ID: "ok-id", Name: "OK", Category: "", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: 1, Quality: 0}}}}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for empty category")
	}
}

func TestValidateGivePacks_ZeroQtyItem(t *testing.T) {
	t.Parallel()
	packs := []givePack{{ID: "ok-id", Name: "OK", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: 0, Quality: 0}}}}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for item with qty=0")
	}
}

func TestValidateGivePacks_NegativeQtyItem(t *testing.T) {
	t.Parallel()
	packs := []givePack{{ID: "ok-id", Name: "OK", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: -1, Quality: 0}}}}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for item with qty=-1")
	}
}

func TestValidateGivePacks_NegativeQualityItem(t *testing.T) {
	t.Parallel()
	packs := []givePack{{ID: "ok-id", Name: "OK", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "A", Qty: 1, Quality: -1}}}}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for item with quality=-1")
	}
}

func TestValidateGivePacks_EmptyTemplateItem(t *testing.T) {
	t.Parallel()
	packs := []givePack{{ID: "ok-id", Name: "OK", Category: "X", Tier: 1, Items: []welcomePackageItem{{Template: "", Qty: 1, Quality: 0}}}}
	if err := validateGivePacks(packs); err == nil {
		t.Fatal("expected error for empty template")
	}
}

func TestValidateGivePacks_EmptyItemsIsValid(t *testing.T) {
	t.Parallel()
	// A pack with zero items is valid — operator may be building it.
	packs := []givePack{{ID: "ok-id", Name: "OK", Category: "X", Tier: 1, Items: []welcomePackageItem{}}}
	if err := validateGivePacks(packs); err != nil {
		t.Fatalf("pack with empty items should be valid: %v", err)
	}
}

// ── parseDefaultPacks ────────────────────────────────────────────────────────

func TestParseDefaultPacks_NonEmpty(t *testing.T) {
	t.Parallel()
	packs, err := parseDefaultPacks()
	if err != nil {
		t.Fatalf("parseDefaultPacks: %v", err)
	}
	if len(packs) == 0 {
		t.Fatal("expected at least one default pack from embedded JSON")
	}
	// Spot-check shape: every pack must have ID, Name, Category.
	for _, p := range packs {
		if p.ID == "" {
			t.Error("pack missing ID")
		}
		if p.Name == "" {
			t.Errorf("pack %q missing name", p.ID)
		}
		if p.Category == "" {
			t.Errorf("pack %q missing category", p.ID)
		}
	}
}

func TestParseDefaultPacks_ValidShape(t *testing.T) {
	t.Parallel()
	packs, err := parseDefaultPacks()
	if err != nil {
		t.Fatalf("parseDefaultPacks: %v", err)
	}
	if err := validateGivePacks(packs); err != nil {
		t.Fatalf("default packs fail validation: %v", err)
	}
}

// ── seedGivePacks ────────────────────────────────────────────────────────────

func TestSeedGivePacks_SetsBasePacksLoaded(t *testing.T) {
	s := setupGivePacksStore(t)

	if err := seedGivePacks(); err != nil {
		t.Fatalf("seedGivePacks: %v", err)
	}

	loaded, packsJSON, ok, err := s.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig after seed: %v", err)
	}
	if !ok {
		t.Fatal("expected config row after seed")
	}
	if !loaded {
		t.Error("expected base_packs_loaded=true after seedGivePacks")
	}
	if packsJSON == "" || packsJSON == "null" || packsJSON == "[]" {
		t.Error("expected non-empty packs JSON after seed")
	}
}

func TestSeedGivePacks_NilStore(t *testing.T) {
	// When the store is nil, seedGivePacks should return an error gracefully.
	givePacksStoreDB = nil
	err := seedGivePacks()
	if err == nil {
		t.Fatal("expected error when store is nil")
	}
}

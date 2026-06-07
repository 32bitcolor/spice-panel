package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// givePacksDefaultJSON is the seed data for first-boot population of give-packs.db.
// Baked into the binary so GoReleaser ships a single self-contained executable.
// To update: edit this literal to match the desired default state, then run make verify.
var givePacksDefaultJSON = []byte(`{"packs":{"t6-starter":{"name":"T6","category":"Starter","tier":6,"items":[{"template":"Combat_Choam_Heavy06_Boots","qty":1,"quality":0},{"template":"Combat_Choam_Heavy06_Gloves","qty":1,"quality":0},{"template":"Combat_Choam_Heavy06_Helmet","qty":1,"quality":0},{"template":"Combat_Choam_Heavy06_Bottom","qty":1,"quality":0},{"template":"Combat_Choam_Heavy06_Top","qty":1,"quality":0},{"template":"Ammo","qty":500,"quality":0},{"template":"HeavyAmmo","qty":500,"quality":0},{"template":"Kindjal_4","qty":1,"quality":0},{"template":"UniqueSword_05","qty":1,"quality":0},{"template":"ChoamSda7","qty":1,"quality":0},{"template":"SmugDmr6","qty":1,"quality":0},{"template":"MiningTool_2h_Advanced","qty":1,"quality":0},{"template":"DewReaper_2h_Tier6","qty":1,"quality":0},{"template":"Literjon_T6","qty":1,"quality":0},{"template":"BodyFluidExtractor_2h_tier6","qty":1,"quality":0},{"template":"Bloodsack_T6","qty":1,"quality":0}]},"sandcrawler-t6":{"name":"T6","category":"Sandcrawler","tier":6,"items":[{"template":"SandcrawlerChassis_6","qty":1,"quality":0},{"template":"SandcrawlerEngine_Unique_Speed_06","qty":1,"quality":0},{"template":"SandcrawlerGenerator_6","qty":1,"quality":0},{"template":"SandcrawlerHull_6","qty":1,"quality":0},{"template":"SandcrawlerLocomotion_Unique_WormThreat_06","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":6,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"t1-starter":{"name":"T1","category":"Starter","tier":1,"items":[{"template":"Combat_Nati_ScavengerRags02_Boots","qty":1,"quality":0},{"template":"Combat_Nati_ScavengerRags02_Gloves","qty":1,"quality":0},{"template":"Combat_Nati_ScavengerRags02_Helmet","qty":1,"quality":0},{"template":"Combat_Nati_ScavengerRags02_Bottom","qty":1,"quality":0},{"template":"Combat_Nati_ScavengerRags02_Top","qty":1,"quality":0},{"template":"Ammo","qty":500,"quality":0},{"template":"HeavyAmmo","qty":500,"quality":0},{"template":"Kindjal","qty":1,"quality":0},{"template":"ChoamSda2","qty":1,"quality":0},{"template":"HarkAr2","qty":1,"quality":0},{"template":"MiningTool_1h_Heavy","qty":1,"quality":0},{"template":"HighCapacityLiterjon","qty":1,"quality":0},{"template":"BodyFluidExtractor","qty":1,"quality":0},{"template":"Bloodsack_01","qty":1,"quality":0}]},"carrier-t6":{"name":"T6","category":"Carrier","tier":6,"items":[{"template":"OrnithopterTransportBoost_Unique_LessHeat_06","qty":1,"quality":0},{"template":"OrnithopterTransportChassis_6","qty":1,"quality":0},{"template":"OrnithopterTransportEngine_6","qty":1,"quality":0},{"template":"OrnithopterTransportGenerator_6","qty":1,"quality":0},{"template":"OrnithopterTransportHull_6","qty":1,"quality":0},{"template":"OrnithopterTransportHullBack_6","qty":1,"quality":0},{"template":"OrnithopterTransportHullFront_6","qty":1,"quality":0},{"template":"OrnithopterTransportLocomotion_Unique_Speed_6","qty":8,"quality":0},{"template":"FuelCanister_Large","qty":6,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"sandbike-t1":{"name":"T1","category":"Sandbike","tier":1,"items":[{"template":"SandbikeChassis_1","qty":1,"quality":0},{"template":"SandbikeEngine_Unique_Speed_1","qty":1,"quality":0},{"template":"SandbikeGenerator_1","qty":1,"quality":0},{"template":"SandbikeHull_1","qty":1,"quality":0},{"template":"SandbikeInventory_1","qty":1,"quality":0},{"template":"SandbikeLocomotion_1","qty":3,"quality":0},{"template":"SandbikeSeat_1","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool","qty":1,"quality":0}]},"sandbike-t6":{"name":"T6","category":"Sandbike","tier":6,"items":[{"template":"SandbikeBoost_Unique_LessHeat_6","qty":1,"quality":0},{"template":"SandbikeChassis_6","qty":1,"quality":0},{"template":"SandbikeEngine_Unique_Speed_6","qty":1,"quality":0},{"template":"SandbikeGenerator_6","qty":1,"quality":0},{"template":"SandbikeHull_6","qty":1,"quality":0},{"template":"SandbikeLocomotion_6","qty":3,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"treadwheel-t5":{"name":"T5","category":"Treadwheel","tier":5,"items":[{"template":"TreadwheelBoost_Unique_LessHeat_5","qty":1,"quality":0},{"template":"TreadwheelChassis_5","qty":1,"quality":0},{"template":"TreadwheelEngine_Unique_Speed_5","qty":1,"quality":0},{"template":"TreadwheelGenerator_5","qty":1,"quality":0},{"template":"TreadwheelLocomotion_5","qty":2,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"assault-t6":{"name":"T6","category":"Assault","tier":6,"items":[{"template":"OrnithopterMediumBoost_Unique_LessHeat_6","qty":1,"quality":0},{"template":"OrnithopterMediumChassis_6","qty":1,"quality":0},{"template":"OrnithopterMediumEngine_6","qty":1,"quality":0},{"template":"OrnithopterMediumGenerator_6","qty":1,"quality":0},{"template":"OrnithopterMediumHull_6","qty":1,"quality":0},{"template":"OrnithopterMediumHullBack_6","qty":1,"quality":0},{"template":"OrnithopterMediumHullFront_6","qty":1,"quality":0},{"template":"OrnithopterMediumLauncher_6","qty":1,"quality":0},{"template":"OrnithopterMediumLocomotion_Unique_Strafe_6","qty":6,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"treadwheel-t6":{"name":"T6","category":"Treadwheel","tier":6,"items":[{"template":"TreadwheelBoost_Unique_LessHeat_6","qty":1,"quality":0},{"template":"TreadwheelChassis_6","qty":1,"quality":0},{"template":"TreadwheelEngine_Unique_Speed_6","qty":1,"quality":0},{"template":"TreadwheelGenerator_6","qty":1,"quality":0},{"template":"TreadwheelLocomotion_6","qty":2,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"t4-starter":{"name":"T4","category":"Starter","tier":4,"items":[{"template":"Combat_Choam_Heavy01_Boots","qty":1,"quality":0},{"template":"Combat_Choam_Heavy01_Gloves","qty":1,"quality":0},{"template":"Combat_Choam_Heavy01_Helmet","qty":1,"quality":0},{"template":"Combat_Choam_Heavy01_Bottom","qty":1,"quality":0},{"template":"Combat_Choam_Heavy01_Top","qty":1,"quality":0},{"template":"Combat_Light_SpiceMask","qty":1,"quality":0},{"template":"Ammo","qty":500,"quality":0},{"template":"HeavyAmmo","qty":500,"quality":0},{"template":"Kindjal_2","qty":1,"quality":0},{"template":"UniqueSword_03","qty":1,"quality":0},{"template":"ChoamSda5","qty":1,"quality":0},{"template":"SmugDmr4","qty":1,"quality":0},{"template":"MiningTool_2h_Heavy","qty":1,"quality":0},{"template":"DewReaper_03","qty":1,"quality":0},{"template":"HighCapacityLiterjon_04","qty":1,"quality":0},{"template":"BodyFluidExtractor_03","qty":1,"quality":0},{"template":"Bloodsack_03","qty":1,"quality":0}]},"buggy-t4":{"name":"T4","category":"Buggy","tier":4,"items":[{"template":"BuggyBoost_Unique_LessHeat_4","qty":1,"quality":0},{"template":"BuggyChassis_4","qty":1,"quality":0},{"template":"BuggyEngine_Unique_Accelerate_04","qty":1,"quality":0},{"template":"BuggyGenerator_4","qty":1,"quality":0},{"template":"BuggyHullBack_4","qty":1,"quality":0},{"template":"BuggyHullBackExtra_4","qty":1,"quality":0},{"template":"BuggyHullFront_4","qty":1,"quality":0},{"template":"BuggyInventory_Unique_Capacity_04","qty":1,"quality":0},{"template":"BuggyLocomotion_4","qty":4,"quality":0},{"template":"BuggyMining_Unique_YieldIncrease_04","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool3","qty":1,"quality":0}]},"scout-t5":{"name":"T5","category":"Scout","tier":5,"items":[{"template":"OrnithopterLightBoost_Unique_LessHeat_5","qty":1,"quality":0},{"template":"OrnithopterLightChassis_5","qty":1,"quality":0},{"template":"OrnithopterLightEngine_5","qty":1,"quality":0},{"template":"OrnithopterLightGenerator_5","qty":1,"quality":0},{"template":"OrnithopterLightHullBack_5","qty":1,"quality":0},{"template":"OrnithopterLightHullFront_5","qty":1,"quality":0},{"template":"OrnithopterLightLauncher_5","qty":1,"quality":0},{"template":"OrnithopterLightLocomotion_Unique_Speed_5","qty":4,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"scout-t6":{"name":"T6","category":"Scout","tier":6,"items":[{"template":"OrnithopterLightBoost_Unique_LessHeat_6","qty":1,"quality":0},{"template":"OrnithopterLightChassis_6","qty":1,"quality":0},{"template":"OrnithopterLightEngine_6","qty":1,"quality":0},{"template":"OrnithopterLightGenerator_6","qty":1,"quality":0},{"template":"OrnithopterLightHullBack_6","qty":1,"quality":0},{"template":"OrnithopterLightHullFront_6","qty":1,"quality":0},{"template":"OrnithopterLightLauncher_6","qty":1,"quality":0},{"template":"OrnithopterLightLocomotion_Unique_Speed_6","qty":4,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"assault-t5":{"name":"T5","category":"Assault","tier":5,"items":[{"template":"OrnithopterMediumBoost_Unique_LessHeat_5","qty":1,"quality":0},{"template":"OrnithopterMediumChassis_5","qty":1,"quality":0},{"template":"OrnithopterMediumEngine_5","qty":1,"quality":0},{"template":"OrnithopterMediumGenerator_5","qty":1,"quality":0},{"template":"OrnithopterMediumHull_5","qty":1,"quality":0},{"template":"OrnithopterMediumHullBack_5","qty":1,"quality":0},{"template":"OrnithopterMediumHullFront_5","qty":1,"quality":0},{"template":"OrnithopterMediumInventory_5","qty":1,"quality":0},{"template":"OrnithopterMediumLauncher_5","qty":1,"quality":0},{"template":"OrnithopterMediumLocomotion_Unique_Strafe_5","qty":6,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"sandbike-t2":{"name":"T2","category":"Sandbike","tier":2,"items":[{"template":"SandbikeBoost_Unique_LessHeat_2","qty":1,"quality":0},{"template":"SandbikeChassis_2","qty":1,"quality":0},{"template":"SandbikeEngine_Unique_Speed_2","qty":1,"quality":0},{"template":"SandbikeGenerator_2","qty":1,"quality":0},{"template":"SandbikeHull_2","qty":1,"quality":0},{"template":"SandbikeInventory_2","qty":1,"quality":0},{"template":"SandbikeLocomotion_2","qty":3,"quality":0},{"template":"SandbikeScanner_2","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool","qty":1,"quality":0}]},"sandbike-t3":{"name":"T3","category":"Sandbike","tier":3,"items":[{"template":"SandbikeBoost_Unique_LessHeat_3","qty":1,"quality":0},{"template":"SandbikeChassis_3","qty":1,"quality":0},{"template":"SandbikeEngine_Unique_Speed_3","qty":1,"quality":0},{"template":"SandbikeGenerator_3","qty":1,"quality":0},{"template":"SandbikeHull_3","qty":1,"quality":0},{"template":"SandbikeLocomotion_3","qty":3,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool3","qty":1,"quality":0}]},"buggy-t6":{"name":"T6","category":"Buggy","tier":6,"items":[{"template":"BuggyBoost_Unique_LessHeat_6","qty":1,"quality":0},{"template":"BuggyChassis_6","qty":1,"quality":0},{"template":"BuggyEngine_Unique_Accelerate_06","qty":1,"quality":0},{"template":"BuggyGenerator_6","qty":1,"quality":0},{"template":"BuggyHullBack_6","qty":1,"quality":0},{"template":"BuggyHullBackExtra_6","qty":1,"quality":0},{"template":"BuggyHullFront_6","qty":1,"quality":0},{"template":"BuggyInventory_Unique_Capacity_06","qty":1,"quality":0},{"template":"BuggyLauncher_6","qty":1,"quality":0},{"template":"BuggyLocomotion_6","qty":4,"quality":0},{"template":"BuggyMining_Unique_YieldIncrease_06","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"scout-t4":{"name":"T4","category":"Scout","tier":4,"items":[{"template":"OrnithopterLightBoost_Unique_LessHeat_4","qty":1,"quality":0},{"template":"OrnithopterLightChassis_4","qty":1,"quality":0},{"template":"OrnithopterLightEngine_4","qty":1,"quality":0},{"template":"OrnithopterLightGenerator_4","qty":1,"quality":0},{"template":"OrnithopterLightHullBack_4","qty":1,"quality":0},{"template":"OrnithopterLightHullFront_4","qty":1,"quality":0},{"template":"OrnithopterLightInventory_4","qty":1,"quality":0},{"template":"OrnithopterLightLocomotion_Unique_Speed_4","qty":4,"quality":0},{"template":"OrnithopterLightScanner_4","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool3","qty":1,"quality":0}]},"sandbike-t4":{"name":"T4","category":"Sandbike","tier":4,"items":[{"template":"SandbikeBoost_Unique_LessHeat_4","qty":1,"quality":0},{"template":"SandbikeChassis_4","qty":1,"quality":0},{"template":"SandbikeEngine_Unique_Speed_4","qty":1,"quality":0},{"template":"SandbikeGenerator_4","qty":1,"quality":0},{"template":"SandbikeHull_4","qty":1,"quality":0},{"template":"SandbikeLocomotion_4","qty":3,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool3","qty":1,"quality":0}]},"t3-starter":{"name":"T3","category":"Starter","tier":3,"items":[{"template":"Combat_Choam_Heavy03_Boots","qty":1,"quality":0},{"template":"Combat_Choam_Heavy03_Gloves","qty":1,"quality":0},{"template":"Combat_Choam_Heavy03_Helmet","qty":1,"quality":0},{"template":"Combat_Choam_Heavy03_Bottom","qty":1,"quality":0},{"template":"Combat_Choam_Heavy03_Top","qty":1,"quality":0},{"template":"Ammo","qty":500,"quality":0},{"template":"HeavyAmmo","qty":500,"quality":0},{"template":"Kindjal_1","qty":1,"quality":0},{"template":"UniqueSword_02","qty":1,"quality":0},{"template":"ChoamSda4","qty":1,"quality":0},{"template":"SmugDmr3","qty":1,"quality":0},{"template":"MiningTool_2h_Standard","qty":1,"quality":0},{"template":"DewReaper_02","qty":1,"quality":0},{"template":"HighCapacityLiterjon_03","qty":1,"quality":0},{"template":"BodyFluidExtractor_02","qty":1,"quality":0},{"template":"Bloodsack_Unique_Durable_03","qty":1,"quality":0}]},"buggy-t3":{"name":"T3","category":"Buggy","tier":3,"items":[{"template":"BuggyBoost_Unique_LessHeat_3","qty":1,"quality":0},{"template":"BuggyChassis_3","qty":1,"quality":0},{"template":"BuggyEngine_Unique_Accelerate_03","qty":1,"quality":0},{"template":"BuggyGenerator_3","qty":1,"quality":0},{"template":"BuggyHullBack_3","qty":1,"quality":0},{"template":"BuggyHullBackExtra_3","qty":1,"quality":0},{"template":"BuggyHullFront_3","qty":1,"quality":0},{"template":"BuggyInventory_Unique_Capacity_03","qty":1,"quality":0},{"template":"BuggyLocomotion_3","qty":4,"quality":0},{"template":"BuggyMining_Unique_YieldIncrease_03","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool3","qty":1,"quality":0}]},"buggy-t5":{"name":"T5","category":"Buggy","tier":5,"items":[{"template":"BuggyBoost_Unique_LessHeat_5","qty":1,"quality":0},{"template":"BuggyChassis_5","qty":1,"quality":0},{"template":"BuggyEngine_Unique_Accelerate_05","qty":1,"quality":0},{"template":"BuggyGenerator_5","qty":1,"quality":0},{"template":"BuggyHullBack_5","qty":1,"quality":0},{"template":"BuggyHullBackExtra_5","qty":1,"quality":0},{"template":"BuggyHullFront_5","qty":1,"quality":0},{"template":"BuggyInventory_Unique_Capacity_05","qty":1,"quality":0},{"template":"BuggyLauncher_5","qty":1,"quality":0},{"template":"BuggyLocomotion_5","qty":4,"quality":0},{"template":"BuggyMining_Unique_YieldIncrease_05","qty":1,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"sandbike-t5":{"name":"T5","category":"Sandbike","tier":5,"items":[{"template":"SandbikeBoost_Unique_LessHeat_5","qty":1,"quality":0},{"template":"SandbikeChassis_5","qty":1,"quality":0},{"template":"SandbikeEngine_Unique_Speed_5","qty":1,"quality":0},{"template":"SandbikeGenerator_5","qty":1,"quality":0},{"template":"SandbikeHull_5","qty":1,"quality":0},{"template":"SandbikeLocomotion_5","qty":3,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool5","qty":1,"quality":0}]},"treadwheel-t4":{"name":"T4","category":"Treadwheel","tier":4,"items":[{"template":"TreadwheelBoost_Unique_LessHeat_4","qty":1,"quality":0},{"template":"TreadwheelChassis_4","qty":1,"quality":0},{"template":"TreadwheelEngine_Unique_Speed_4","qty":1,"quality":0},{"template":"TreadwheelGenerator_4","qty":1,"quality":0},{"template":"TreadwheelLocomotion_4","qty":2,"quality":0},{"template":"FuelCanister_Large","qty":2,"quality":0},{"template":"WeldingMaterial","qty":500,"quality":0},{"template":"RepairTool3","qty":1,"quality":0}]},"t5-starter":{"name":"T5","category":"Starter","tier":5,"items":[{"template":"Combat_Choam_Heavy04_Shoes","qty":1,"quality":0},{"template":"Combat_Choam_Heavy04_Gloves","qty":1,"quality":0},{"template":"Combat_Choam_Heavy04_Helmet","qty":1,"quality":0},{"template":"Combat_Choam_Heavy04_Bottom","qty":1,"quality":0},{"template":"Combat_Choam_Heavy04_Top","qty":1,"quality":0},{"template":"Combat_Light_SpiceMask","qty":1,"quality":0},{"template":"Ammo","qty":500,"quality":0},{"template":"HeavyAmmo","qty":500,"quality":0},{"template":"Kindjal_3","qty":1,"quality":0},{"template":"UniqueSword_04","qty":1,"quality":0},{"template":"ChoamSda6","qty":1,"quality":0},{"template":"SmugDmr5","qty":1,"quality":0},{"template":"MiningTool_2h_Light","qty":1,"quality":0},{"template":"DewReaper_2h_Unique_YieldIncrease_05","qty":1,"quality":0},{"template":"HighCapacityLiterjon_05","qty":1,"quality":0},{"template":"BodyFluidExtractor_Unique_Water_05","qty":1,"quality":0},{"template":"Bloodsack_Unique_Durable_05","qty":1,"quality":0}]}}}`)

// givePacksStoreDB is the global SQLite store for operator-configured packs.
// Set once by initGivePacksStore in main.go; nil when the store failed to open.
var givePacksStoreDB *givePacksStore

// givePack is one configurable give-items pack entry. The operator can CRUD
// these via the Manage Packages modal. IDs are stable, unique string keys
// (e.g. "starter-t1") that map to entries in the packs dropdown.
type givePack struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	Category string               `json:"category"`
	Tier     int                  `json:"tier"`
	Items    []welcomePackageItem `json:"items"`
}

// cdnPackEntry mirrors the shape in give_packs_default.json / CDN packs.json.
// The top-level key becomes givePack.ID.
type cdnPackEntry struct {
	Name     string               `json:"name"`
	Category string               `json:"category"`
	Tier     int                  `json:"tier"`
	Items    []welcomePackageItem `json:"items"`
}

type cdnPacksFile struct {
	Packs map[string]cdnPackEntry `json:"packs"`
}

// parseDefaultPacks parses the embedded give_packs_default.json (CDN format)
// into a []givePack slice sorted deterministically by ID.
func parseDefaultPacks() ([]givePack, error) {
	var raw cdnPacksFile
	if err := json.Unmarshal(givePacksDefaultJSON, &raw); err != nil {
		return nil, fmt.Errorf("parse default packs: %w", err)
	}
	packs := make([]givePack, 0, len(raw.Packs))
	for id, entry := range raw.Packs {
		items := entry.Items
		if items == nil {
			items = []welcomePackageItem{}
		}
		packs = append(packs, givePack{
			ID:       id,
			Name:     entry.Name,
			Category: entry.Category,
			Tier:     entry.Tier,
			Items:    items,
		})
	}
	return packs, nil
}

// validateGivePacks returns an error if any pack fails basic consistency rules.
// An empty slice is valid (operator deleted all packs intentionally).
func validateGivePacks(packs []givePack) error {
	seen := make(map[string]bool, len(packs))
	for i, p := range packs {
		if strings.TrimSpace(p.ID) == "" {
			return fmt.Errorf("pack at index %d: id must not be empty", i)
		}
		if seen[p.ID] {
			return fmt.Errorf("pack %q: duplicate id", p.ID)
		}
		seen[p.ID] = true
		if strings.TrimSpace(p.Name) == "" {
			return fmt.Errorf("pack %q: name must not be empty", p.ID)
		}
		if strings.TrimSpace(p.Category) == "" {
			return fmt.Errorf("pack %q: category must not be empty", p.ID)
		}
		if err := validateGivePackItems(p.ID, p.Items); err != nil {
			return err
		}
	}
	return nil
}

// validateGivePackItems validates the items within a single pack.
// An empty items slice is allowed (operator building a pack progressively).
func validateGivePackItems(packID string, items []welcomePackageItem) error {
	for _, it := range items {
		if strings.TrimSpace(it.Template) == "" {
			return fmt.Errorf("pack %q: item template must not be empty", packID)
		}
		if it.Qty <= 0 {
			return fmt.Errorf("pack %q: item %q qty must be > 0", packID, it.Template)
		}
		if it.Quality < 0 {
			return fmt.Errorf("pack %q: item %q quality must be >= 0", packID, it.Template)
		}
	}
	return nil
}

// seedGivePacks parses the embedded default packs and persists them into the
// store with base_packs_loaded=true. Called once at startup when the store row
// is missing (ok=false) or base_packs_loaded=false. Never called again, so
// user edits — including deleting all packs — are never overwritten.
func seedGivePacks() error {
	if givePacksStoreDB == nil {
		return fmt.Errorf("give-packs store not available")
	}
	packs, err := parseDefaultPacks()
	if err != nil {
		return fmt.Errorf("seed give packs: %w", err)
	}
	packsJSON, err := json.Marshal(packs)
	if err != nil {
		return fmt.Errorf("seed give packs marshal: %w", err)
	}
	return givePacksStoreDB.saveConfig(string(packsJSON), true)
}

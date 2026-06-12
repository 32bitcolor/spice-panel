package main

import (
	"encoding/json"
	"fmt"
)

// Default battlepass catalog: ~1,469 bonus intel across 158 tiers, vs the
// 2,779 intel a character earns naturally from levels 1–200. Tier keys are
// stable identities ("level:5", "journey:DA_MQ_FindTheFremen",
// "tag:Exploration.Ecolab.HotIce") so the catalog can be reseeded without
// losing claim history. Node IDs and tags were extracted from the game data
// dump (dune-item-data: MnemonicRecall journey cards + ExplorationTags).

// bpItems encodes one-of-each item templates as a tier RewardItems JSON
// string. Quality 0 — schematics are not gradeable; runGiveItem writes the
// stack directly to the offline player's inventory.
func bpItems(templates ...string) string {
	items := make([]rewardItem, len(templates))
	for i, tpl := range templates {
		items[i] = rewardItem{Template: tpl, Qty: 1, Quality: 0}
	}
	b, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(b)
}

// battlepassLevelRewards maps level tiers to unique-schematic rewards
// (template IDs from the game data dump). The tier of the schematic ramps
// with the level band; milestone levels (50/100/150/200) award full sets.
var battlepassLevelRewards = map[int64]string{
	5:  bpItems("Schematic_UniqueCutteray2"),           // Sim's Cutter (T1)
	10: bpItems("Schematic_UniqueLiterjon"),            // Hajra Literjon Mk1 (T1)
	15: bpItems("Schematic_UniqueMaulaPistol"),         // Way of the Fallen (T1)
	20: bpItems("PowerPack_Unique_Regen_01_Schematic"), // Old Sparky Mk1 (T1)
	25: bpItems("Kindjal_Unique_Blood_01_Schematic"),   // Kaleff's Drinker (T1)
	30: bpItems("Schematic_UniqueSuspensor"),           // The Emperor's Wings Mk1 (T1)
	35: bpItems("UniqueAr_Burst_01_Schematic"),         // Aren's Vengeance (T1)
	40: bpItems("Schematic_UniqueScattergun"),          // Shredder (T2)
	45: bpItems("Schematic_UniqueChoamSword"),          // Pseudo-Pulse-Sword (T2)
	50: bpItems( // Aren's light set (T1)
		"Schematic_UniquePincushionHead", "Schematic_UniquePincushionChest",
		"Schematic_UniquePincushionHands", "Schematic_UniquePincushionLegs",
		"Schematic_UniquePincushionFeet"),
	55: bpItems("Schematic_UniqueCutteray3"),           // Olef's Quickcutter (T2)
	60: bpItems("SMG_Unique_LargeMag_02_Schematic"),    // Legion Tattoo (T2)
	65: bpItems("Kindjal_Unique_Blood_02_Schematic"),   // Scipio's Drinker (T2)
	70: bpItems("PowerPack_Unique_Regen_02_Schematic"), // Old Sparky Mk2 (T2)
	75: bpItems("Schematic_UniqueDewReaper"),           // Buoyant Reaper Mk2 (T2)
	80: bpItems("Schematic_UniqueBattleRifle"),         // The Tapper (T3)
	85: bpItems("Schematic_UniqueDirk"),                // Searing Shiv (T3)
	90: bpItems("SMG_Unique_LargeMag_03_Schematic"),    // Seb's Kisser (T3)
	95: bpItems("UniqueSword_02_Schematic"),            // Shock-sword (T3)
	100: bpItems( // Oathbreaker heavy set (T2)
		"Schematic_UniqueOathbreakerHead", "Schematic_UniqueOathbreakerChest",
		"Schematic_UniqueOathbreakerHands", "Schematic_UniqueOathbreakerLegs",
		"Schematic_UniqueOathbreakerFeet"),
	105: bpItems("Schematic_UniqueCutteray4"),             // Callie's Breaker (T3)
	110: bpItems("HeavyPistol_Unique_Bleed_03_Schematic"), // Artisan Disruptor Pistol (T3)
	115: bpItems("LongRifle_Unique_Poison_03_Schematic"),  // Assassin's Rifle (T3)
	120: bpItems("Kindjal_Unique_Stamina_03_Schematic"),   // Shock-Knife (T3)
	125: bpItems("UniqueScattergun2_Schematic"),           // Ripper (T3)
	130: bpItems("PowerPack_Unique_Regen_03_Schematic"),   // Old Sparky Mk3 (T3)
	135: bpItems("Scanner_Unique_Body_03_Schematic"),      // Handheld Life Scanner Mk3 (T3)
	140: bpItems("UniqueAr_Burst_03_Schematic"),           // Zaal's Companion (T3)
	145: bpItems("UniqueSda3_Schematic"),                  // Way of the Lost (T3)
	150: bpItems( // Karak's heavy set (T3)
		"Combat_Neut_AtreidesDeserterUnique02_Helmet_Schematic",
		"Combat_Neut_AtreidesDeserterUnique02_Top_Schematic",
		"Combat_Neut_AtreidesDeserterUnique02_Gloves_Schematic",
		"Combat_Neut_AtreidesDeserterUnique02_Bottom_Schematic",
		"Combat_Neut_AtreidesDeserterUnique02_Boots_Schematic"),
	155: bpItems("Schematic_UniqueCutteray5"),                // Tarl Cutteray (T5)
	160: bpItems("SMG_Unique_LargeMag_05_Schematic"),         // Relentless (T5)
	165: bpItems("HeavyPistol_Unique_Headshot_05_Schematic"), // Cope (T5)
	170: bpItems("Kindjal_Unique_Blood_05_Schematic"),        // Pardot's Drinker (T5)
	175: bpItems("LMG_Unique_RapidFire_05_Schematic"),        // Extravagant Message (T5)
	180: bpItems("Schematic_UniqueFlamethrower"),             // Shaitan's Tongue (T5)
	185: bpItems("LongRifle_Unique_LargeMag_05_Schematic"),   // Fivefinger's Tripleshot Rifle (T5)
	190: bpItems("Shotgun_Unique_Explosive_05_Schematic"),    // Adept Burst Drillshot (T5)
	195: bpItems("PowerPack_Unique_Regen_05_Schematic"),      // Young Sparky Mk5 (T5)
	200: bpItems( // Bulwark heavy set + Branding Blade (T6)
		"Combat_Heavy_Unique_Reinforced_Helmet_06_Schematic",
		"Combat_Heavy_Unique_Reinforced_Top_06_Schematic",
		"Combat_Heavy_Unique_Reinforced_Gloves_06_Schematic",
		"Combat_Heavy_Unique_Reinforced_Bottom_06_Schematic",
		"Combat_Heavy_Unique_Reinforced_Boots_06_Schematic",
		"B1C4_Unique_Sword2_Schematic"),
}

// battlepassLevelIntel returns the intel reward for a level tier.
func battlepassLevelIntel(level int64) int64 {
	switch {
	case level >= 200:
		return 75
	case level > 150:
		return 25
	case level > 100:
		return 20
	case level > 50:
		return 15
	default:
		return 10
	}
}

// battlepassStoryChapters are the main quest journey card root nodes.
// DA_MQ_NPEAutocompleted (tutorial) is rewarded lightly. Item rewards are
// themed unique schematics that ramp with the chapter order.
var battlepassStoryChapters = []struct {
	node  string
	label string
	intel int64
	items string
}{
	{"DA_MQ_NPEAutocompleted", "Tutorial: Arrival on Arrakis", 10,
		bpItems("Stillsuit_Unique_Armored_01_Mask_Schematic")}, // Hollower Mask (T1)
	{"DA_MQ_FindTheFremen", "Chapter: Find the Fremen", 40,
		bpItems( // rest of the Hollower stillsuit (T1)
			"Stillsuit_Unique_Armored_01_Top_Schematic",
			"Stillsuit_Unique_Armored_01_Gloves_Schematic",
			"Stillsuit_Unique_Armored_01_Boots_Schematic")},
	{"DA_MQ_AssassinsHandbook", "Chapter: The Assassin's Handbook", 40,
		bpItems("Kindjal_Unique_Blood_03_Schematic")}, // Glutton's Drinker (T3)
	{"DA_MQ_TheGreatConvention", "Chapter: The Great Convention", 40,
		bpItems("Combat_Light_Unique_BiomeHeat_Top_03_Schematic")}, // Skin-lined Jacket (T3)
	{"DA_MQ_TheGreatConventionPt2", "Chapter: The Great Convention Pt. 2", 40,
		bpItems("Combat_Heavy_Unique_StandStillDmgReduction_Top_06_Schematic")}, // Fortress Chestpiece
	{"DA_MQ_ANewBeginning", "Chapter: A New Beginning", 40,
		bpItems( // Saturnine stillsuit (T5)
			"Stillsuit_Unique_Armored_05_Mask_Schematic",
			"Stillsuit_Unique_Armored_05_Gloves_Schematic",
			"Stillsuit_Unique_Armored_05_Boots_Schematic")},
	{"DA_MQ_TheBloodline", "Chapter: The Bloodline", 40,
		bpItems( // B1C4 chapter uniques (T6)
			"B1C4_Unique_Rapier2_Schematic",        // Prescient Edge
			"B1C4_Unique_HeavyPistol2_Schematic")}, // Hell-Fury Pistol
}

// battlepassSideQuests are the regional side quest journey card roots, each
// paired with a region-flavored unique schematic (mostly T3–T5 utility).
var battlepassSideQuests = []struct {
	node  string
	label string
	items string
}{
	{"DA_SQ_OverlandMap", "Side Quest: Overland Map",
		bpItems("SandbikeScanner_Unique_LongRange_02_Schematic")}, // Duneman Sandbike Scanner Mk2
	{"DA_SQ_DeepDesert", "Side Quest: Deep Desert",
		bpItems("Combat_Light_Unique_WormThreat_Boots_05_Schematic")}, // Tarl Softstep Boots
	{"DA_SQ_Sheol", "Side Quest: Sheol",
		bpItems("Combat_Neut_SmugglerDeserterUnique03_Top_Schematic")}, // Inkvine Jacket
	{"DA_SQ_Oodham", "Side Quest: O'odham",
		bpItems("DewReaper_Unique_02_Schematic")}, // Buoyant Reaper Mk3
	{"DA_SQ_Taxation", "Side Quest: Taxation",
		bpItems("StaticCompactor_Unique_Compact_03_Schematic")}, // Compact Compactor Mk3
	{"DA_SQ_JabalEifrit", "Side Quest: Jabal Eifrit",
		bpItems("Combat_Light_Unique_BiomeHeat_Top_05_Schematic")}, // Station Garb
	{"DA_SQ_VermiliusGap", "Side Quest: Vermilius Gap",
		bpItems("BodyFluidExtractor_Unique_Water_03_Schematic")}, // Filter Extractor Mk3
}

// battlepassExplorationTags are player tags granted on discovery. Item
// rewards: altars build Kel's stillsuit + softstep boots, ecolabs award
// science/hydration uniques, trade posts pay schematic pattern fragments,
// shipwrecks build Mendia's smuggler set, POIs award faction set pieces.
var battlepassExplorationTags = []struct {
	tag   string
	label string
	intel int64
	items string
}{
	{"Exploration.Cave.Large.Altar1", "Discover Cave Altar 1", 5,
		bpItems("Stillsuit_Unique_Armored_03_Mask_Schematic")}, // Kel's Mask
	{"Exploration.Cave.Large.Altar2", "Discover Cave Altar 2", 5,
		bpItems("Stillsuit_Unique_Armored_03_Gloves_Schematic")}, // Kel's Gloves
	{"Exploration.Cave.Large.Altar3", "Discover Cave Altar 3", 5,
		bpItems("Stillsuit_Unique_Armored_03_Boots_Schematic")}, // Kel's Boots
	{"Exploration.Cave.Large.Altar4", "Discover Cave Altar 4", 5,
		bpItems("Stillsuit_Unique_Armored_03_Top_Schematic")}, // Kel's Garment
	{"Exploration.Cave.Large.Altar5", "Discover Cave Altar 5", 5,
		bpItems("Combat_Light_Unique_WormThreat_Boots_02_Schematic")}, // Softstep Boots
	{"Exploration.Cave.Large.Altar6", "Discover Cave Altar 6", 5,
		bpItems("Combat_Light_Unique_WormThreat_Boots_03_Schematic")}, // Ta'lab Softstep Boots
	{"Exploration.Cave.Large.Altar7", "Discover Cave Altar 7", 5,
		bpItems("Combat_Light_Unique_WormThreat_Boots_06_Schematic")}, // Tabr Softstep Boots
	{"Exploration.Ecolab.BondageAndBotany", "Ecolab: Bondage and Botany", 5,
		bpItems("BodyFluidExtractor_Unique_Poison_05_Schematic")}, // Impure Extractor Mk5
	{"Exploration.Ecolab.HotIce", "Ecolab: Hot Ice", 5,
		bpItems("BodyFluidExtractor_Unique_Water_04_Schematic")}, // Filter Extractor Mk4
	{"Exploration.Ecolab.RedScorpion", "Ecolab: Red Scorpion", 5,
		bpItems("Bloodsack_Unique_Durable_03_Schematic")}, // Glutton's Bloodbag
	{"Exploration.Ecolab.TheBlight", "Ecolab: The Blight", 5,
		bpItems("Combat_Light_Unique_DewReap_Gloves_03_Schematic")}, // Reaper Gloves
	{"Exploration.Ecolab.TheBoosters", "Ecolab: The Boosters", 5,
		bpItems("Combat_Light_Unique_Stamina_Bottom_05_Schematic")}, // Stim-Leggings
	{"Exploration.Ecolab.TheHijacker", "Ecolab: The Hijacker", 5,
		bpItems("DewReaper_1h_Unique_Compact_05_Schematic")}, // Collapsible Dew Reaper Mk5
	{"Exploration.Ecolab.WaterTrap", "Ecolab: Water Trap", 5,
		bpItems("DewReaper_2h_Unique_YieldIncrease_05_Schematic")}, // Focused Reaper Mk5
	{"Exploration.POI.TradePost.BreakersYard", "Trade Post: Breaker's Yard", 3,
		bpItems("T6SchematicFragmentQL1")}, // Schematic Pattern Grade 1
	{"Exploration.POI.TradePost.Crossroads", "Trade Post: Crossroads", 3,
		bpItems("T6SchematicFragmentQL2")}, // Schematic Pattern Grade 2
	{"Exploration.POI.TradePost.GriffinsReach", "Trade Post: Griffin's Reach", 3,
		bpItems("T6SchematicFragmentQL3")}, // Schematic Pattern Grade 3
	{"Exploration.POI.TradePost.PinnacleStation", "Trade Post: Pinnacle Station", 3,
		bpItems("T6SchematicFragmentQL4")}, // Schematic Pattern Grade 4
	{"Exploration.POI.TradePost.TheAnvil", "Trade Post: The Anvil", 3,
		bpItems("T6SchematicFragmentQL5")}, // Schematic Pattern Grade 5
	{"Exploration.Shipwreck.Medium.HaggaRift", "Shipwreck: Hagga Rift", 5,
		bpItems("Combat_Neut_SmugglerDeserterUnique02_Helmet_Schematic")}, // Mendia's Wrap
	{"Exploration.Shipwreck.Medium.JabalEast", "Shipwreck: Jabal East", 5,
		bpItems("Combat_Neut_SmugglerDeserterUnique02_Top_Schematic")}, // Mendia's Jacket
	{"Exploration.Shipwreck.Medium.Jimbob", "Shipwreck: Jimbob", 5,
		bpItems("Combat_Neut_SmugglerDeserterUnique02_Gloves_Schematic")}, // Mendia's Gauntlets
	{"Exploration.Shipwreck.Medium.Sheol", "Shipwreck: Sheol", 5,
		bpItems("Combat_Neut_SmugglerDeserterUnique02_Bottom_Schematic")}, // Mendia's Pants
	{"Exploration.Shipwreck.Medium.ShieldWallWest", "Shipwreck: Shield Wall West", 5,
		bpItems("Combat_Neut_SmugglerDeserterUnique02_Boots_Schematic")}, // Mendia's Boots
	{"Exploration.Shipwreck.PolarCap.NML", "Shipwreck: Polar Cap", 5,
		bpItems("Combat_Light_Unique_Scanning_Helmet_05_Schematic")}, // Wayfinder Helm
	{"Exploration.POI.Arrakeen", "Discover Arrakeen", 5,
		bpItems("BodyFluidExtractor_Unique_Water_05_Schematic")}, // Filter Extractor Mk5
	{"Exploration.POI.HarkoVillage", "Discover Harko Village", 5,
		bpItems("Combat_Hark_MedUnique02_Helmet_Schematic")}, // The Forge Helmet
	{"Exploration.POI.Fortress.Atreides", "Discover the Atreides Fortress", 5,
		bpItems("Combat_Neut_AtreidesDeserterUnique04_Helmet_Schematic")}, // Acheronian Helmet
	{"Exploration.POI.Fortress.Harkonnen", "Discover the Harkonnen Fortress", 5,
		bpItems("Combat_Hark_MedUnique02_Top_Schematic")}, // The Forge Chestpiece
	{"Exploration.POI.Landmark.Mirzabah", "Landmark: Mirzabah", 5,
		bpItems("Combat_Light_Unique_Climbing_Gloves_06_Schematic")}, // Hook-claw Gloves
	{"Exploration.POI.Landmark.SpyNest", "Landmark: Spy Nest", 5,
		bpItems("Combat_Light_Unique_StaminaDmgIncrease_Helmet_06_Schematic")}, // Seeker Helmet
}

// battlepassAchievementSlugs are the child node names under
// DA_ACH_SteamAchievements (note: there is no social3 in the game data).
var battlepassAchievementSlugs = map[string][]int{
	"combat":      {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	"exploration": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	"industry":    {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
	"misc":        {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	"progression": {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
	"skills":      {1, 2, 3, 4, 5, 6, 7, 8, 9},
	"social":      {1, 2, 4, 5},
}

// battlepassAchievementCategories fixes iteration order for deterministic IDs.
var battlepassAchievementCategories = []string{
	"combat", "exploration", "industry", "misc", "progression", "skills", "social",
}

// defaultBattlepassCatalog builds the seed tier list.
func defaultBattlepassCatalog() []battlepassTier {
	var out []battlepassTier

	for level := int64(5); level <= 200; level += 5 {
		out = append(out, battlepassTier{
			TierKey:     fmt.Sprintf("level:%d", level),
			Category:    "level",
			Label:       fmt.Sprintf("Reach Level %d", level),
			Signal:      battlepassSignalLevel,
			Threshold:   level,
			Intel:       battlepassLevelIntel(level),
			RewardItems: battlepassLevelRewards[level],
			Enabled:     true,
		})
	}

	for _, ch := range battlepassStoryChapters {
		out = append(out, battlepassTier{
			TierKey:     "journey:" + ch.node,
			Category:    "story",
			Label:       ch.label,
			Signal:      battlepassSignalJourneyNode,
			SignalKey:   ch.node,
			Intel:       ch.intel,
			RewardItems: ch.items,
			Enabled:     true,
		})
	}

	for _, sq := range battlepassSideQuests {
		out = append(out, battlepassTier{
			TierKey:     "journey:" + sq.node,
			Category:    "side_quest",
			Label:       sq.label,
			Signal:      battlepassSignalJourneyNode,
			SignalKey:   sq.node,
			Intel:       20,
			RewardItems: sq.items,
			Enabled:     true,
		})
	}

	out = append(out, battlepassTier{
		TierKey:   "journey:DA_FQ_ClimbTheRanks",
		Category:  "faction",
		Label:     "Faction: Climb the Ranks",
		Signal:    battlepassSignalJourneyNode,
		SignalKey: "DA_FQ_ClimbTheRanks",
		Intel:     40,
		RewardItems: bpItems( // Mendek's heavy set (T5) — house operator regalia
			"Combat_Heavy_Unique_Reinforced_Helmet_05_Schematic",
			"Combat_Heavy_Unique_Reinforced_Top_05_Schematic",
			"Combat_Heavy_Unique_Reinforced_Gloves_05_Schematic",
			"Combat_Heavy_Unique_Reinforced_Bottom_05_Schematic",
			"Combat_Heavy_Unique_Reinforced_Boots_05_Schematic"),
		Enabled: true,
	})

	for _, ex := range battlepassExplorationTags {
		out = append(out, battlepassTier{
			TierKey:     "tag:" + ex.tag,
			Category:    "exploration",
			Label:       ex.label,
			Signal:      battlepassSignalPlayerTag,
			SignalKey:   ex.tag,
			Intel:       ex.intel,
			RewardItems: ex.items,
			Enabled:     true,
		})
	}

	for _, cat := range battlepassAchievementCategories {
		for _, n := range battlepassAchievementSlugs[cat] {
			node := fmt.Sprintf("DA_ACH_SteamAchievements.sb-ach-%s%d", cat, n)
			out = append(out, battlepassTier{
				TierKey:   "journey:" + node,
				Category:  "achievement",
				Label:     fmt.Sprintf("Achievement: %s %d", battlepassTitleCase(cat), n),
				Signal:    battlepassSignalJourneyNode,
				SignalKey: node,
				Intel:     2,
				Enabled:   true,
			})
		}
	}

	return out
}

// battlepassTitleCase upper-cases the first ASCII letter of a category slug.
func battlepassTitleCase(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 'a' - 'A'
	}
	return string(b)
}

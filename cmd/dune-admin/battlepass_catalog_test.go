package main

import (
	"encoding/json"
	"testing"
)

func TestBattlepassCatalogShape(t *testing.T) {
	catalog := defaultBattlepassCatalog()

	counts := map[string]int{}
	totals := map[string]int64{}
	keys := map[string]bool{}
	var total int64
	for _, tier := range catalog {
		if keys[tier.TierKey] {
			t.Errorf("duplicate tier key %q", tier.TierKey)
		}
		keys[tier.TierKey] = true
		counts[tier.Category]++
		totals[tier.Category] += tier.Intel
		total += tier.Intel

		if !tier.Enabled {
			t.Errorf("tier %q must default enabled", tier.TierKey)
		}
		if tier.Intel <= 0 {
			t.Errorf("tier %q has non-positive intel %d", tier.TierKey, tier.Intel)
		}
		switch tier.Signal {
		case battlepassSignalLevel:
			if tier.Threshold < 1 || tier.Threshold > 200 {
				t.Errorf("level tier %q threshold %d out of range", tier.TierKey, tier.Threshold)
			}
		case battlepassSignalJourneyNode, battlepassSignalPlayerTag:
			if tier.SignalKey == "" {
				t.Errorf("tier %q missing signal key", tier.TierKey)
			}
		default:
			t.Errorf("tier %q has unknown signal %q", tier.TierKey, tier.Signal)
		}
	}

	want := map[string]int{
		"level":       40,
		"story":       7,
		"side_quest":  7,
		"faction":     31,
		"exploration": 31,
		"achievement": 72,
	}
	for cat, n := range want {
		if counts[cat] != n {
			t.Errorf("category %s has %d tiers, want %d", cat, counts[cat], n)
		}
	}
	if len(catalog) != 188 {
		t.Errorf("catalog has %d tiers, want 188", len(catalog))
	}

	// Budget sanity: 1,469 base + 150 new contract tiers = 1,619.
	if total != 1619 {
		t.Errorf("catalog total intel = %d, want 1619", total)
	}
	if totals["level"] != 750 {
		t.Errorf("level track total = %d, want 750", totals["level"])
	}
}

func TestBattlepassCatalogRewardItems(t *testing.T) {
	seen := map[string]string{}
	for _, tier := range defaultBattlepassCatalog() {
		if tier.RewardItems == "" {
			continue
		}
		var items []rewardItem
		if err := json.Unmarshal([]byte(tier.RewardItems), &items); err != nil {
			t.Errorf("tier %s reward_items is not valid JSON: %v", tier.TierKey, err)
			continue
		}
		if len(items) == 0 {
			t.Errorf("tier %s has empty reward_items array (use \"\" instead)", tier.TierKey)
		}
		for _, item := range items {
			if item.Template == "" || item.Qty < 1 {
				t.Errorf("tier %s has invalid reward item %+v", tier.TierKey, item)
			}
			if prev, dup := seen[item.Template]; dup {
				t.Errorf("template %s rewarded by both %s and %s", item.Template, prev, tier.TierKey)
			}
			seen[item.Template] = tier.TierKey
		}
	}
	if len(seen) == 0 {
		t.Fatal("default catalog has no item rewards")
	}
}

func TestBattlepassLevelIntel(t *testing.T) {
	cases := []struct {
		level int64
		want  int64
	}{
		{5, 10}, {50, 10}, {55, 15}, {100, 15},
		{105, 20}, {150, 20}, {155, 25}, {195, 25}, {200, 75},
	}
	for _, c := range cases {
		if got := battlepassLevelIntel(c.level); got != c.want {
			t.Errorf("battlepassLevelIntel(%d) = %d, want %d", c.level, got, c.want)
		}
	}
}

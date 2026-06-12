package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// mockBattlepassDeps returns deps backed by simple in-memory fixtures.
func mockBattlepassDeps(players []battlepassPlayer, journey map[int64][]string, tags map[int64][]string) battlepassDeps {
	return battlepassDeps{
		fetchPlayers: func(ctx context.Context) ([]battlepassPlayer, error) {
			return players, nil
		},
		fetchCompletedJourneyNodes: func(ctx context.Context, accountID int64) ([]string, error) {
			return journey[accountID], nil
		},
		fetchPlayerTags: func(ctx context.Context, accountID int64) ([]string, error) {
			return tags[accountID], nil
		},
	}
}

func engineTestTiers() []battlepassTier {
	return []battlepassTier{
		{TierKey: "level:5", Category: "level", Label: "Level 5", Signal: battlepassSignalLevel, Threshold: 5, Intel: 10, Enabled: true},
		{TierKey: "level:10", Category: "level", Label: "Level 10", Signal: battlepassSignalLevel, Threshold: 10, Intel: 10, Enabled: true},
		{TierKey: "journey:DA_MQ_FindTheFremen", Category: "story", Label: "Find the Fremen", Signal: battlepassSignalJourneyNode, SignalKey: "DA_MQ_FindTheFremen", Intel: 40, Enabled: true},
		{TierKey: "tag:Exploration.Cave.Large.Altar1", Category: "exploration", Label: "Altar 1", Signal: battlepassSignalPlayerTag, SignalKey: "Exploration.Cave.Large.Altar1", Intel: 5, Enabled: true},
		{TierKey: "level:15", Category: "level", Label: "Level 15", Signal: battlepassSignalLevel, Threshold: 15, Intel: 10, Enabled: false},
	}
}

func seededEngineStore(t *testing.T) *battlepassStore {
	t.Helper()
	s := testBattlepassStore(t)
	if _, err := s.seedTiersIfEmpty(engineTestTiers()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return s
}

func TestBattlepassTierSatisfied(t *testing.T) {
	journey := map[string]bool{"DA_MQ_FindTheFremen": true}
	tags := map[string]bool{"Exploration.Cave.Large.Altar1": true}
	tiers := engineTestTiers()

	if !battlepassTierSatisfied(tiers[0], 7, journey, tags) {
		t.Error("level 7 must satisfy level:5")
	}
	if battlepassTierSatisfied(tiers[1], 7, journey, tags) {
		t.Error("level 7 must not satisfy level:10")
	}
	if !battlepassTierSatisfied(tiers[2], 1, journey, tags) {
		t.Error("completed journey node must satisfy journey tier")
	}
	if !battlepassTierSatisfied(tiers[3], 1, journey, tags) {
		t.Error("player tag must satisfy tag tier")
	}
	if battlepassTierSatisfied(tiers[2], 1, map[string]bool{}, tags) {
		t.Error("missing journey node must not satisfy journey tier")
	}
}

func TestBattlepassFirstEvaluationBaselines(t *testing.T) {
	s := seededEngineStore(t)
	players := []battlepassPlayer{{AccountID: 1, PawnID: 100, Name: "Paul", Level: 7}}
	deps := mockBattlepassDeps(players,
		map[int64][]string{1: {"DA_MQ_FindTheFremen"}},
		map[int64][]string{})

	if err := evaluateBattlepassTick(context.Background(), deps, s, false); err != nil {
		t.Fatalf("tick: %v", err)
	}

	keys, _ := s.claimedKeys(1)
	if keys["level:5"] != battlepassClaimBaseline {
		t.Errorf("level:5 = %q, want baseline (pre-existing progress)", keys["level:5"])
	}
	if keys["journey:DA_MQ_FindTheFremen"] != battlepassClaimBaseline {
		t.Errorf("journey claim = %q, want baseline", keys["journey:DA_MQ_FindTheFremen"])
	}
	if _, ok := keys["level:10"]; ok {
		t.Error("unsatisfied tier must not be claimed")
	}
	if _, ok := keys["level:15"]; ok {
		t.Error("disabled tier must not be claimed")
	}
	totals, _ := s.earnedTotals()
	if len(totals) != 0 {
		t.Errorf("baseline run must not create earned intel: %v", totals)
	}
}

func TestBattlepassNewUnlocksEarnAfterBaseline(t *testing.T) {
	s := seededEngineStore(t)
	players := []battlepassPlayer{{AccountID: 1, PawnID: 100, Name: "Paul", Level: 7}}
	journey := map[int64][]string{}
	tags := map[int64][]string{}
	deps := mockBattlepassDeps(players, journey, tags)

	// First tick baselines level:5.
	if err := evaluateBattlepassTick(context.Background(), deps, s, false); err != nil {
		t.Fatalf("first tick: %v", err)
	}

	// Player progresses: level 12, completes the chapter, discovers the altar.
	players[0].Level = 12
	journey[1] = []string{"DA_MQ_FindTheFremen"}
	tags[1] = []string{"Exploration.Cave.Large.Altar1"}

	if err := evaluateBattlepassTick(context.Background(), deps, s, false); err != nil {
		t.Fatalf("second tick: %v", err)
	}

	keys, _ := s.claimedKeys(1)
	if keys["level:5"] != battlepassClaimBaseline {
		t.Errorf("level:5 must stay baseline, got %q", keys["level:5"])
	}
	for _, k := range []string{"level:10", "journey:DA_MQ_FindTheFremen", "tag:Exploration.Cave.Large.Altar1"} {
		if keys[k] != battlepassClaimEarned {
			t.Errorf("%s = %q, want earned", k, keys[k])
		}
	}
	totals, _ := s.earnedTotals()
	if totals[1] != 10+40+5 {
		t.Errorf("earned total = %d, want 55", totals[1])
	}
}

func TestBattlepassAwardPastSkipsBaseline(t *testing.T) {
	s := seededEngineStore(t)
	players := []battlepassPlayer{{AccountID: 1, PawnID: 100, Name: "Paul", Level: 7}}
	deps := mockBattlepassDeps(players, map[int64][]string{}, map[int64][]string{})

	if err := evaluateBattlepassTick(context.Background(), deps, s, true); err != nil {
		t.Fatalf("tick: %v", err)
	}

	keys, _ := s.claimedKeys(1)
	if keys["level:5"] != battlepassClaimEarned {
		t.Errorf("award-past mode: level:5 = %q, want earned", keys["level:5"])
	}
}

func TestBattlepassFetchErrorSkipsBaseline(t *testing.T) {
	s := seededEngineStore(t)
	players := []battlepassPlayer{{AccountID: 1, PawnID: 100, Name: "Paul", Level: 7}}
	deps := mockBattlepassDeps(players, map[int64][]string{}, map[int64][]string{})
	deps.fetchCompletedJourneyNodes = func(ctx context.Context, accountID int64) ([]string, error) {
		return nil, fmt.Errorf("db down")
	}

	// Evaluation fails mid-pass: the account must NOT be marked baselined,
	// so the next successful pass still baselines instead of over-rewarding.
	if err := evaluateBattlepassTick(context.Background(), deps, s, false); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if baselined, _ := s.isBaselined(1); baselined {
		t.Fatal("account must not be baselined after a failed evaluation pass")
	}

	deps.fetchCompletedJourneyNodes = func(ctx context.Context, accountID int64) ([]string, error) {
		return []string{"DA_MQ_FindTheFremen"}, nil
	}
	if err := evaluateBattlepassTick(context.Background(), deps, s, false); err != nil {
		t.Fatalf("second tick: %v", err)
	}
	keys, _ := s.claimedKeys(1)
	if keys["journey:DA_MQ_FindTheFremen"] != battlepassClaimBaseline {
		t.Errorf("journey claim = %q, want baseline", keys["journey:DA_MQ_FindTheFremen"])
	}
}

func TestBattlepassSkipsSignalFetchWhenAllClaimed(t *testing.T) {
	s := seededEngineStore(t)
	players := []battlepassPlayer{{AccountID: 1, PawnID: 100, Name: "Paul", Level: 200}}
	journeyCalls, tagCalls := 0, 0
	deps := battlepassDeps{
		fetchPlayers: func(ctx context.Context) ([]battlepassPlayer, error) { return players, nil },
		fetchCompletedJourneyNodes: func(ctx context.Context, accountID int64) ([]string, error) {
			journeyCalls++
			return []string{"DA_MQ_FindTheFremen"}, nil
		},
		fetchPlayerTags: func(ctx context.Context, accountID int64) ([]string, error) {
			tagCalls++
			return []string{"Exploration.Cave.Large.Altar1"}, nil
		},
	}

	if err := evaluateBattlepassTick(context.Background(), deps, s, false); err != nil {
		t.Fatalf("first tick: %v", err)
	}
	if journeyCalls != 1 || tagCalls != 1 {
		t.Fatalf("first tick fetches = %d/%d, want 1/1", journeyCalls, tagCalls)
	}

	// Everything is claimed — second tick must not re-fetch per-player signals.
	if err := evaluateBattlepassTick(context.Background(), deps, s, false); err != nil {
		t.Fatalf("second tick: %v", err)
	}
	if journeyCalls != 1 || tagCalls != 1 {
		t.Fatalf("second tick re-fetched signals (%d/%d), want no new fetches", journeyCalls, tagCalls)
	}
}

func TestClampBattlepassInterval(t *testing.T) {
	if got := clampBattlepassInterval(0); got != 60*time.Second {
		t.Errorf("default interval = %v, want 60s", got)
	}
	if got := clampBattlepassInterval(5); got != 10*time.Second {
		t.Errorf("clamped low = %v, want 10s", got)
	}
	if got := clampBattlepassInterval(10000); got != 600*time.Second {
		t.Errorf("clamped high = %v, want 600s", got)
	}
	if got := clampBattlepassInterval(120); got != 120*time.Second {
		t.Errorf("interval = %v, want 120s", got)
	}
}

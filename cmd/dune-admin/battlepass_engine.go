package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// battlepassPlayer is the per-character snapshot the battlepass engine needs.
// Level is derived in the bulk fetch so evaluation makes no per-player level
// queries.
type battlepassPlayer struct {
	AccountID int64
	PawnID    int64
	Name      string
	Online    bool
	Level     int
}

// battlepassDeps holds injectable functions so the engine can be unit-tested
// without a live DB (pattern: eventDeps).
type battlepassDeps struct {
	fetchPlayers               func(ctx context.Context) ([]battlepassPlayer, error)
	fetchCompletedJourneyNodes func(ctx context.Context, accountID int64) ([]string, error)
	fetchPlayerTags            func(ctx context.Context, accountID int64) ([]string, error)
}

// globalBattlepassStore is set once at startup, guarded in every handler.
var globalBattlepassStore *battlepassStore

// battlepassTierSatisfied reports whether the player state meets the tier's
// signal condition.
func battlepassTierSatisfied(t battlepassTier, level int, journey, tags map[string]bool) bool {
	switch t.Signal {
	case battlepassSignalLevel:
		return int64(level) >= t.Threshold
	case battlepassSignalJourneyNode:
		return journey[t.SignalKey]
	case battlepassSignalPlayerTag:
		return tags[t.SignalKey]
	default:
		return false
	}
}

// battlepassUnclaimed filters tiers to enabled ones the account has no claim
// for, and reports which signals the remaining tiers need.
func battlepassUnclaimed(tiers []battlepassTier, claimed map[string]string) (unclaimed []battlepassTier, needsJourney, needsTags bool) {
	for _, t := range tiers {
		if !t.Enabled {
			continue
		}
		if _, ok := claimed[t.TierKey]; ok {
			continue
		}
		unclaimed = append(unclaimed, t)
		switch t.Signal {
		case battlepassSignalJourneyNode:
			needsJourney = true
		case battlepassSignalPlayerTag:
			needsTags = true
		}
	}
	return unclaimed, needsJourney, needsTags
}

// evaluateBattlepassPlayer records claims for every newly-satisfied tier.
// During the account's first complete pass (and unless awardPast is set),
// satisfied tiers are recorded as baseline — pre-existing progress is not
// rewarded; the pass pays for new unlocks only.
func evaluateBattlepassPlayer(ctx context.Context, deps battlepassDeps, store *battlepassStore, tiers []battlepassTier, p battlepassPlayer, awardPast bool) error {
	claimed, err := store.claimedKeys(p.AccountID)
	if err != nil {
		return fmt.Errorf("claimed keys: %w", err)
	}
	unclaimed, needsJourney, needsTags := battlepassUnclaimed(tiers, claimed)

	baselined := true
	if !awardPast {
		baselined, err = store.isBaselined(p.AccountID)
		if err != nil {
			return fmt.Errorf("baselined check: %w", err)
		}
	}

	if len(unclaimed) > 0 {
		status := battlepassClaimEarned
		if !baselined {
			status = battlepassClaimBaseline
		}
		if err := recordSatisfiedBattlepassTiers(ctx, deps, store, unclaimed, p, status, needsJourney, needsTags); err != nil {
			return err
		}
	}

	// Only mark baselined after a fully successful pass — failing earlier
	// keeps old progress eligible for baselining, never for rewards.
	if !baselined {
		if err := store.markBaselined(p.AccountID); err != nil {
			return fmt.Errorf("mark baselined: %w", err)
		}
	}
	return nil
}

// recordSatisfiedBattlepassTiers fetches the needed signal sets and records a
// claim with the given status for every satisfied tier.
func recordSatisfiedBattlepassTiers(ctx context.Context, deps battlepassDeps, store *battlepassStore, tiers []battlepassTier, p battlepassPlayer, status string, needsJourney, needsTags bool) error {
	journey, tags, err := fetchBattlepassSignals(ctx, deps, p.AccountID, needsJourney, needsTags)
	if err != nil {
		return err
	}
	for _, t := range tiers {
		if !battlepassTierSatisfied(t, p.Level, journey, tags) {
			continue
		}
		if err := store.recordClaim(t.TierKey, p.AccountID, t.Intel, status); err != nil {
			return fmt.Errorf("record claim %s: %w", t.TierKey, err)
		}
	}
	return nil
}

// fetchBattlepassSignals loads the per-player signal sets that are actually
// needed by unclaimed tiers.
func fetchBattlepassSignals(ctx context.Context, deps battlepassDeps, accountID int64, needsJourney, needsTags bool) (journey, tags map[string]bool, err error) {
	journey, tags = map[string]bool{}, map[string]bool{}
	if needsJourney {
		nodes, err := deps.fetchCompletedJourneyNodes(ctx, accountID)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch journey nodes: %w", err)
		}
		for _, n := range nodes {
			journey[n] = true
		}
	}
	if needsTags {
		tagList, err := deps.fetchPlayerTags(ctx, accountID)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch player tags: %w", err)
		}
		for _, tag := range tagList {
			tags[tag] = true
		}
	}
	return journey, tags, nil
}

// evaluateBattlepassTick evaluates every tracked player against the enabled
// tiers. Per-player failures are logged and skipped so one bad row cannot
// stall the whole pass.
func evaluateBattlepassTick(ctx context.Context, deps battlepassDeps, store *battlepassStore, awardPast bool) error {
	tiers, err := store.listTiers()
	if err != nil {
		return fmt.Errorf("battlepass list tiers: %w", err)
	}
	if len(tiers) == 0 {
		return nil
	}
	players, err := deps.fetchPlayers(ctx)
	if err != nil {
		return fmt.Errorf("battlepass fetch players: %w", err)
	}
	for _, p := range players {
		if err := evaluateBattlepassPlayer(ctx, deps, store, tiers, p, awardPast); err != nil {
			log.Printf("battlepass: evaluate account %d: %v", p.AccountID, err)
		}
	}
	return nil
}

// ── polling loop ──────────────────────────────────────────────────────────────

// clampBattlepassInterval converts BattlepassPollSeconds to a Duration,
// defaulting to 60s and clamped [10s, 600s].
func clampBattlepassInterval(secs int) time.Duration {
	if secs < 1 {
		secs = 60
	}
	if secs < 10 {
		secs = 10
	}
	if secs > 600 {
		secs = 600
	}
	return time.Duration(secs) * time.Second
}

// runBattlepassEngine runs the evaluation loop until ctx is cancelled.
func runBattlepassEngine(ctx context.Context, deps battlepassDeps, store *battlepassStore, interval time.Duration, awardPast bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := evaluateBattlepassTick(ctx, deps, store, awardPast); err != nil {
				log.Printf("battlepass: tick: %v", err)
			}
		}
	}
}

// productionBattlepassDeps builds the deps from live globals. Called from
// startBattlepassIfEnabled only; tests inject mocks directly.
func productionBattlepassDeps() battlepassDeps {
	return battlepassDeps{
		fetchPlayers: func(ctx context.Context) ([]battlepassPlayer, error) {
			if globalDB == nil {
				return nil, fmt.Errorf("database not connected")
			}
			return cmdFetchBattlepassPlayers(ctx, globalDB)
		},
		fetchCompletedJourneyNodes: func(ctx context.Context, accountID int64) ([]string, error) {
			if globalDB == nil {
				return nil, fmt.Errorf("database not connected")
			}
			return cmdFetchCompletedJourneyNodeIDs(ctx, globalDB, accountID)
		},
		fetchPlayerTags: func(ctx context.Context, accountID int64) ([]string, error) {
			if globalDB == nil {
				return nil, fmt.Errorf("database not connected")
			}
			return cmdFetchPlayerTagsForAccount(ctx, globalDB, accountID)
		},
	}
}

// battlepassEnabled returns the effective enabled flag (default off).
func battlepassEnabled(cfg appConfig) bool {
	if cfg.BattlepassEnabled == nil {
		return false
	}
	return *cfg.BattlepassEnabled
}

// battlepassAwardPast returns the award-past flag (default false: the pass
// rewards new unlocks only; pre-existing progress is baselined unrewarded).
func battlepassAwardPast(cfg appConfig) bool {
	if cfg.BattlepassAwardPast == nil {
		return false
	}
	return *cfg.BattlepassAwardPast
}

// startBattlepassIfEnabled seeds the catalog and starts the evaluation loop
// when battlepass_enabled is set. The store is opened either way so the
// admin UI can browse/edit the catalog before enabling.
func startBattlepassIfEnabled(cfg appConfig) {
	if globalBattlepassStore == nil {
		return
	}
	if n, err := globalBattlepassStore.seedTiersIfEmpty(defaultBattlepassCatalog()); err != nil {
		log.Printf("battlepass: seed catalog: %v", err)
	} else if n > 0 {
		log.Printf("battlepass: seeded %d default tiers", n)
	}
	if !battlepassEnabled(cfg) {
		return
	}
	interval := clampBattlepassInterval(cfg.BattlepassPollSeconds)
	log.Printf("battlepass: engine enabled (interval %s, award_past=%t)", interval, battlepassAwardPast(cfg))
	go runBattlepassEngine(context.Background(), productionBattlepassDeps(), globalBattlepassStore, interval, battlepassAwardPast(cfg))
}

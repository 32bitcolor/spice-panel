package main

import "testing"

// The presence tracker detects join events by diffing the online set across
// ticks. Its defining behaviour: the FIRST observation seeds a silent baseline
// (so a dune-admin restart never re-fires on-join actions for everyone already
// in-game), and a player who leaves and returns counts as a fresh join.

func TestPresenceTracker_BaselineThenJoin(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()

	// First observation seeds the baseline — no joins even though two players
	// are already online.
	if joins := p.observe([]welcomeAccount{
		{AccountID: 1, CharacterName: "Paul"},
		{AccountID: 2, CharacterName: "Chani"},
	}); len(joins) != 0 {
		t.Fatalf("baseline tick: want 0 joins, got %d (%+v)", len(joins), joins)
	}

	// Same set again → steady state, no joins.
	if joins := p.observe([]welcomeAccount{
		{AccountID: 1, CharacterName: "Paul"},
		{AccountID: 2, CharacterName: "Chani"},
	}); len(joins) != 0 {
		t.Fatalf("steady state: want 0 joins, got %d (%+v)", len(joins), joins)
	}

	// A new player appears (account 3) → exactly one join.
	joins := p.observe([]welcomeAccount{
		{AccountID: 1, CharacterName: "Paul"},
		{AccountID: 3, CharacterName: "Leto"},
	})
	if len(joins) != 1 || joins[0].AccountID != 3 {
		t.Fatalf("new player: want 1 join for acct 3, got %+v", joins)
	}
}

func TestPresenceTracker_RejoinAfterLeave(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	p.observe([]welcomeAccount{{AccountID: 1}}) // baseline

	// Player 1 goes offline → no join.
	if joins := p.observe(nil); len(joins) != 0 {
		t.Fatalf("after leave: want 0 joins, got %+v", joins)
	}
	// Player 1 returns → counts as a fresh join.
	joins := p.observe([]welcomeAccount{{AccountID: 1}})
	if len(joins) != 1 || joins[0].AccountID != 1 {
		t.Fatalf("rejoin: want 1 join for acct 1, got %+v", joins)
	}
}

func TestPresenceTracker_Empty(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	if joins := p.observe(nil); len(joins) != 0 {
		t.Fatalf("empty baseline: want 0 joins, got %+v", joins)
	}
	if joins := p.observe([]welcomeAccount{}); len(joins) != 0 {
		t.Fatalf("empty steady: want 0 joins, got %+v", joins)
	}
}

// observeJoinsLeaves diffs the online set across ticks, returning BOTH joins
// (newly online) and leaves (gone since the last tick). Like observe(), the very
// first call seeds a silent baseline so a dune-admin restart never re-fires
// join/leave actions for everyone already in-game.

func TestPresenceTracker_JoinsLeaves_Baseline(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	joins, leaves := p.observeJoinsLeaves([]welcomeAccount{
		{AccountID: 1, CharacterName: "Paul", Region: "HaggaBasin"},
		{AccountID: 2, CharacterName: "Chani", Region: "HaggaBasin"},
	})
	if len(joins) != 0 || len(leaves) != 0 {
		t.Fatalf("baseline tick: want 0 joins/0 leaves, got joins=%+v leaves=%+v", joins, leaves)
	}
}

func TestPresenceTracker_JoinsLeaves_Join(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	p.observeJoinsLeaves([]welcomeAccount{{AccountID: 1, CharacterName: "Paul", Region: "HaggaBasin"}})

	joins, leaves := p.observeJoinsLeaves([]welcomeAccount{
		{AccountID: 1, CharacterName: "Paul", Region: "HaggaBasin"},
		{AccountID: 2, CharacterName: "Chani", Region: "TheShield"},
	})
	if len(leaves) != 0 {
		t.Fatalf("want 0 leaves, got %+v", leaves)
	}
	if len(joins) != 1 || joins[0].AccountID != 2 || joins[0].Region != "TheShield" {
		t.Fatalf("want 1 join for acct 2 in TheShield, got %+v", joins)
	}
}

func TestPresenceTracker_JoinsLeaves_LeaveRetainsNameRegion(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	p.observeJoinsLeaves([]welcomeAccount{
		{AccountID: 1, CharacterName: "Paul", Region: "HaggaBasin"},
		{AccountID: 2, CharacterName: "Chani", Region: "TheShield"},
	})

	// Account 2 goes offline; the leave event must carry its last-known name and
	// region (the new snapshot no longer includes it).
	joins, leaves := p.observeJoinsLeaves([]welcomeAccount{
		{AccountID: 1, CharacterName: "Paul", Region: "HaggaBasin"},
	})
	if len(joins) != 0 {
		t.Fatalf("want 0 joins, got %+v", joins)
	}
	if len(leaves) != 1 {
		t.Fatalf("want 1 leave, got %+v", leaves)
	}
	if leaves[0].AccountID != 2 || leaves[0].CharacterName != "Chani" || leaves[0].Region != "TheShield" {
		t.Fatalf("leave should retain last-known name/region, got %+v", leaves[0])
	}
}

func TestPresenceTracker_JoinsLeaves_SimultaneousJoinAndLeave(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	p.observeJoinsLeaves([]welcomeAccount{{AccountID: 1, CharacterName: "Paul", Region: "HaggaBasin"}})

	joins, leaves := p.observeJoinsLeaves([]welcomeAccount{
		{AccountID: 2, CharacterName: "Chani", Region: "TheShield"},
	})
	if len(joins) != 1 || joins[0].AccountID != 2 {
		t.Fatalf("want 1 join for acct 2, got %+v", joins)
	}
	if len(leaves) != 1 || leaves[0].AccountID != 1 {
		t.Fatalf("want 1 leave for acct 1, got %+v", leaves)
	}
}

// observe() and observeJoinsLeaves() share the same baseline state. Mixing them
// must not double-fire: this guards the MOTD path (observe) and the region path
// (observeJoinsLeaves) running off one tracker.
func TestPresenceTracker_ObserveSharesBaselineWithJoinsLeaves(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	if joins := p.observe([]welcomeAccount{{AccountID: 1}}); len(joins) != 0 {
		t.Fatalf("observe baseline: want 0 joins, got %+v", joins)
	}
	// Already seeded — observeJoinsLeaves should not treat acct 1 as a fresh join.
	joins, leaves := p.observeJoinsLeaves([]welcomeAccount{{AccountID: 1}})
	if len(joins) != 0 || len(leaves) != 0 {
		t.Fatalf("seeded: want 0/0, got joins=%+v leaves=%+v", joins, leaves)
	}
}

// reset() re-arms the baseline: the next observe is silent (used when MOTD is
// toggled off then on, so currently-online players aren't messaged on the flip).
func TestPresenceTracker_Reset(t *testing.T) {
	t.Parallel()
	p := newPresenceTracker()
	p.observe([]welcomeAccount{{AccountID: 1}}) // baseline with player 1

	p.reset()
	// Post-reset, the next tick is a fresh baseline — player 2 is NOT a join.
	if joins := p.observe([]welcomeAccount{{AccountID: 1}, {AccountID: 2}}); len(joins) != 0 {
		t.Fatalf("post-reset baseline: want 0 joins, got %+v", joins)
	}
	// Subsequent new players are detected again.
	joins := p.observe([]welcomeAccount{{AccountID: 1}, {AccountID: 2}, {AccountID: 3}})
	if len(joins) != 1 || joins[0].AccountID != 3 {
		t.Fatalf("post-reset join: want acct 3, got %+v", joins)
	}
}

package main

// presenceTracker detects player join (and leave) events by diffing the set of
// online accounts across successive observations. The first observation seeds a
// silent baseline so a dune-admin (re)start does not re-fire on-join/on-leave
// actions (e.g. the MOTD or region broadcasts) for everyone already in-game; a
// player who goes offline and returns is a new join. Keyed on account id (always
// present), which is also what the whisper path consumes. The full last-known
// welcomeAccount is retained per id so leave events can carry the player's name
// and region after they have dropped out of the online snapshot.
//
// Not safe for concurrent use: the scanner calls observe()/observeJoinsLeaves()
// serially from its single goroutine.
type presenceTracker struct {
	seen   map[int64]welcomeAccount
	seeded bool
}

func newPresenceTracker() *presenceTracker {
	return &presenceTracker{seen: map[int64]welcomeAccount{}}
}

// observe records the currently-online accounts and returns those newly online
// since the previous observation (join events). The first call returns no joins
// (it only seeds the baseline).
func (p *presenceTracker) observe(online []welcomeAccount) []welcomeAccount {
	joins, _ := p.diff(online)
	return joins
}

// observeJoinsLeaves records the currently-online accounts and returns BOTH the
// newly-online accounts (joins) and the accounts gone since the previous
// observation (leaves). Leave events carry the last-known welcomeAccount (name +
// region) captured on the prior snapshot, since the current snapshot no longer
// includes them. The first call returns no joins/leaves (it only seeds).
func (p *presenceTracker) observeJoinsLeaves(online []welcomeAccount) (joins, leaves []welcomeAccount) {
	return p.diff(online)
}

// diff updates the tracked online set to the given snapshot and returns the
// joins and leaves relative to the prior snapshot. Shared by observe and
// observeJoinsLeaves so both consume one consistent baseline.
func (p *presenceTracker) diff(online []welcomeAccount) (joins, leaves []welcomeAccount) {
	current := make(map[int64]welcomeAccount, len(online))
	for _, acc := range online {
		current[acc.AccountID] = acc
		if !p.seeded {
			continue
		}
		if _, ok := p.seen[acc.AccountID]; !ok {
			joins = append(joins, acc)
		}
	}
	if p.seeded {
		for id, prev := range p.seen {
			if _, ok := current[id]; !ok {
				leaves = append(leaves, prev)
			}
		}
	}
	p.seen = current
	p.seeded = true
	return joins, leaves
}

// reset re-arms the baseline so the next observe is silent. Used when the MOTD
// feature is toggled off (and later on) so currently-online players are not
// messaged on the flip — only genuine future joins are.
func (p *presenceTracker) reset() {
	p.seen = map[int64]welcomeAccount{}
	p.seeded = false
}

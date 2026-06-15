package main

import (
	"context"
	"time"
)

// warmerInterval is the refresh-ahead cadence. Kept below healthCacheTTL so a
// steadily-polled entry is repopulated before it expires and UI reads always
// hit a warm cache.
const warmerInterval = 10 * time.Second

// prewarmTimeout bounds the boot prewarm so a slow/unreachable control plane
// can't delay startup (assembleServerHealth has its own 8s GetStatus timeout
// per server).
const prewarmTimeout = 30 * time.Second

// cacheWarmer proactively refreshes hot read-cache entries before they expire
// (refresh-ahead) and prewarms them at boot, so the UI is served from cache
// instead of triggering a cold fan-out on every poll / hard refresh.
type cacheWarmer struct {
	registry *serverRegistry
	interval time.Duration
}

func newCacheWarmer(reg *serverRegistry) *cacheWarmer {
	return &cacheWarmer{registry: reg, interval: warmerInterval}
}

// warmAll refreshes every warmable entry for every registered server. Currently
// just per-server health; broaden as more reads are cached.
func (w *cacheWarmer) warmAll(ctx context.Context) {
	if globalHealthCache == nil || w.registry == nil {
		return
	}
	for _, sc := range w.registry.All() {
		h := assembleServerHealth(ctx, sc)
		globalHealthCache.Set(cacheKey(sc.ID, "health"), h, healthCacheTTL)
	}
}

// run refreshes on a fixed cadence until ctx is cancelled (process shutdown).
func (w *cacheWarmer) run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.warmAll(ctx)
		}
	}
}

// prewarmCaches populates hot caches once at boot so the first UI paint is
// instant — no cold-miss fan-out on a hard refresh. Best-effort + time-bounded.
func prewarmCaches(parent context.Context, w *cacheWarmer) {
	ctx, cancel := context.WithTimeout(parent, prewarmTimeout)
	defer cancel()
	w.warmAll(ctx)
}

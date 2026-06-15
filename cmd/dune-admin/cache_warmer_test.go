package main

import (
	"context"
	"testing"
)

func TestCacheWarmer_WarmAllPopulatesHealth(t *testing.T) {
	origCache := globalHealthCache
	c, err := newRistrettoCache[serverHealth]("test-health", 256)
	if err != nil {
		t.Fatalf("newRistrettoCache: %v", err)
	}
	globalHealthCache = c
	t.Cleanup(func() { globalHealthCache = origCache })

	ctrl := &healthFakeControl{status: &BattlegroupStatus{Phase: "Running", Database: "Connected"}}
	reg := newServerRegistry(nil)
	reg.Register(&ServerContext{ID: "1", Name: "One", Cfg: ServerConfig{ID: 1, Control: "amp"}, Control: ctrl})
	reg.Register(&ServerContext{ID: "2", Name: "Two", Cfg: ServerConfig{ID: 2, Control: "amp"}, Control: ctrl})

	// Cold cache → both miss.
	if _, ok := c.Get(cacheKey("1", "health")); ok {
		t.Fatal("expected cold miss before warm")
	}

	newCacheWarmer(reg).warmAll(context.Background())

	for _, scope := range []string{"1", "2"} {
		h, ok := c.Get(cacheKey(scope, "health"))
		if !ok {
			t.Errorf("scope %s: health not warmed", scope)
			continue
		}
		if !h.Running {
			t.Errorf("scope %s: warmed health Running=false, want true", scope)
		}
	}
}

func TestPrewarmCaches_PopulatesBeforeFirstRequest(t *testing.T) {
	origCache := globalHealthCache
	c, err := newRistrettoCache[serverHealth]("test-health", 256)
	if err != nil {
		t.Fatalf("newRistrettoCache: %v", err)
	}
	globalHealthCache = c
	t.Cleanup(func() { globalHealthCache = origCache })

	ctrl := &healthFakeControl{status: &BattlegroupStatus{Phase: "Running", Database: "Connected"}}
	reg := newServerRegistry(nil)
	reg.Register(&ServerContext{ID: "1", Name: "One", Cfg: ServerConfig{ID: 1, Control: "amp"}, Control: ctrl})

	prewarmCaches(context.Background(), newCacheWarmer(reg))

	if _, ok := c.Get(cacheKey("1", "health")); !ok {
		t.Error("prewarm did not populate health cache")
	}
}

func TestCacheWarmer_NoCacheIsNoop(t *testing.T) {
	origCache := globalHealthCache
	globalHealthCache = nil
	t.Cleanup(func() { globalHealthCache = origCache })

	reg := newServerRegistry(nil)
	reg.Register(&ServerContext{ID: "1", Name: "One"})
	// Must not panic when the cache is unavailable.
	newCacheWarmer(reg).warmAll(context.Background())
}

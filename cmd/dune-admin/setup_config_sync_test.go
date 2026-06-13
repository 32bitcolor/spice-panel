package main

import "testing"

// TestWriteSetupConfigSyncsLoadedConfig guards the fix for the embedded market
// bot starting on a fresh install even after the operator answered "no": the
// setup wizard wrote the config file but left the in-memory loadedConfig at its
// zero value, where MarketBotEnabled==nil defaults to ON. writeSetupConfig must
// sync loadedConfig so startEmbeddedMarketBotIfEnabled sees the chosen value.
func TestWriteSetupConfigSyncsLoadedConfig(t *testing.T) {
	t.Setenv("DUNE_ADMIN_CONFIG_DIR", t.TempDir())
	prev := loadedConfig
	t.Cleanup(func() { loadedConfig = prev })

	disabled := false
	cfg := appConfig{Control: "kubectl", MarketBotEnabled: &disabled}

	writeSetupConfig(func(string) {}, func(string) {}, cfg)

	if loadedConfig.MarketBotEnabled == nil {
		t.Fatal("loadedConfig.MarketBotEnabled is nil after setup — would default the bot ON")
	}
	if *loadedConfig.MarketBotEnabled {
		t.Fatal("loadedConfig.MarketBotEnabled = true after the operator chose 'no'")
	}
	if marketBotEnabled(loadedConfig) {
		t.Fatal("marketBotEnabled(loadedConfig) = true — embedded bot would start despite 'no'")
	}
}

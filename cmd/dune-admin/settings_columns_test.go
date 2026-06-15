package main

import "testing"

func TestSettingsColumnsRoundTrip(t *testing.T) {
	db := openSharedScopeDB(t)
	tru := true
	in := globalSettingsOnly(appConfig{
		ListenAddr: ":8080", ScripCurrency: 7, ServerIniDir: "/ini",
		DBPass:                      "should-be-cleared-by-globalSettingsOnly", // connection field → cleared
		BrokerPass:                  "bpw",
		BrokerJWTSecret:             "jwt",
		AmpAPIPass:                  "amppw",
		AmpUseContainer:             &tru,
		AmpAPIPort:                  8081,
		MarketBotEnabled:            &tru,
		MarketBotThresh:             1.5,
		MarketBotRemoteToken:        "tok",
		DiscordBotToken:             "dtok",
		DiscordStatusEnabled:        &tru,
		AuthEnabled:                 &tru,
		AuthLocalUsername:           "admin",
		AuthLocalPasswordHash:       "hash",
		AuthDiscordClientSecret:     "secret",
		AuthSessionTTLHours:         24,
		BattlepassEnabled:           &tru,
		BattlepassPollSeconds:       60,
		WelcomePackageEnabled:       &tru,
		WelcomePackageActiveVersion: "v2",
		EventsEnabled:               &tru,
	})
	if err := saveSettingsColumns(db, in); err != nil {
		t.Fatalf("saveSettingsColumns: %v", err)
	}
	got, ok, err := loadSettingsColumns(db)
	if err != nil || !ok {
		t.Fatalf("loadSettingsColumns: ok=%v err=%v", ok, err)
	}
	// Secrets survive.
	if got.BrokerPass != "bpw" || got.AmpAPIPass != "amppw" ||
		got.AuthLocalPasswordHash != "hash" || got.AuthDiscordClientSecret != "secret" ||
		got.MarketBotRemoteToken != "tok" || got.DiscordBotToken != "dtok" || got.BrokerJWTSecret != "jwt" {
		t.Errorf("secret lost in round-trip: %+v", got)
	}
	// Tri-state *bool survives as true.
	for name, p := range map[string]*bool{
		"AuthEnabled": got.AuthEnabled, "AmpUseContainer": got.AmpUseContainer,
		"MarketBotEnabled": got.MarketBotEnabled, "DiscordStatusEnabled": got.DiscordStatusEnabled,
		"BattlepassEnabled": got.BattlepassEnabled, "WelcomePackageEnabled": got.WelcomePackageEnabled,
		"EventsEnabled": got.EventsEnabled,
	} {
		if p == nil || !*p {
			t.Errorf("%s *bool lost (want true, got %v)", name, p)
		}
	}
	// Numerics + strings.
	if got.MarketBotThresh != 1.5 || got.AuthSessionTTLHours != 24 || got.AmpAPIPort != 8081 ||
		got.ScripCurrency != 7 || got.ListenAddr != ":8080" || got.WelcomePackageActiveVersion != "v2" {
		t.Errorf("scalar lost: %+v", got)
	}
}

func TestSettingsColumns_BoolPtrFalseAndNil(t *testing.T) {
	db := openSharedScopeDB(t)
	fls := false
	if err := saveSettingsColumns(db, appConfig{AuthEnabled: &fls}); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, _, err := loadSettingsColumns(db)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.AuthEnabled == nil || *got.AuthEnabled {
		t.Errorf("AuthEnabled = %v, want explicit false", got.AuthEnabled)
	}
	// MarketBotEnabled was never set → must round-trip as nil (unset), not false.
	if got.MarketBotEnabled != nil {
		t.Errorf("MarketBotEnabled = %v, want nil (unset)", got.MarketBotEnabled)
	}
}

func TestMigrateSettingsColumns(t *testing.T) {
	db := openSharedScopeDB(t)
	// Seed the OLD app_settings.config_json blob.
	old := appConfig{ListenAddr: ":7000", AuthLocalUsername: "old", AuthLocalPasswordHash: "h"}
	if err := newSettingsStore(db).saveSettingsBlob(old); err != nil {
		t.Fatalf("seed blob: %v", err)
	}
	if err := migrateSettingsColumns(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	got, ok, err := loadSettingsColumns(db)
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if got.AuthLocalUsername != "old" || got.AuthLocalPasswordHash != "h" || got.ListenAddr != ":7000" {
		t.Errorf("migration lost data: %+v", got)
	}
	if m, _ := metaGet(db, "migrated:settings_columns"); m == "" {
		t.Error("marker not set")
	}

	// Idempotent + blob ignored: mutate the stale blob, re-run, reads stay migrated.
	if err := newSettingsStore(db).saveSettingsBlob(appConfig{ListenAddr: ":9999"}); err != nil {
		t.Fatalf("re-seed blob: %v", err)
	}
	if err := migrateSettingsColumns(db); err != nil {
		t.Fatalf("re-migrate: %v", err)
	}
	got2, _, _ := loadSettingsColumns(db)
	if got2.ListenAddr != ":7000" {
		t.Errorf("post-migration read followed stale blob: ListenAddr=%q", got2.ListenAddr)
	}
}

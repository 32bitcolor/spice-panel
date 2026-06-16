package main

import (
	"runtime"
	"testing"
)

func TestBuildEnvironmentAllowlist(t *testing.T) {
	origCfg := loadedConfig
	enabled := true
	loadedConfig = appConfig{Control: "amp", AuthEnabled: &enabled, MarketBotEnabled: &enabled}
	t.Cleanup(func() { loadedConfig = origCfg })

	env := buildEnvironment()
	if env.ControlPlane != "amp" {
		t.Errorf("ControlPlane = %q, want amp", env.ControlPlane)
	}
	if !env.AuthEnabled || !env.MarketBot {
		t.Errorf("expected auth + market bot enabled, got %+v", env)
	}
	if env.GoVersion != runtime.Version() || env.OS != runtime.GOOS {
		t.Errorf("runtime fields wrong: %+v", env)
	}
	if env.Version != AppVersion {
		t.Errorf("Version = %q, want %q", env.Version, AppVersion)
	}
}

func TestBuildEnvironmentControlDefault(t *testing.T) {
	origCfg := loadedConfig
	loadedConfig = appConfig{} // blank control
	t.Cleanup(func() { loadedConfig = origCfg })
	if got := buildEnvironment().ControlPlane; got != "local" {
		t.Errorf("blank control should default to local, got %q", got)
	}
}

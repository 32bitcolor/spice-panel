package main

import "runtime"

// environmentSummary is the allowlist-only environment block included in
// diagnostics artifacts. Adding a field is a deliberate code change — nothing
// is emitted that is not named here.
type environmentSummary struct {
	Version      string `json:"version"`
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	ControlPlane string `json:"control_plane"`
	AuthEnabled  bool   `json:"auth_enabled"`
	MarketBot    bool   `json:"market_bot_enabled"`
	ServerCount  int    `json:"active_server_count"`
}

func buildEnvironment() environmentSummary {
	return environmentSummary{
		Version:      AppVersion,
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		ControlPlane: controlOrDefault(loadedConfig.Control),
		AuthEnabled:  authEnabled(loadedConfig),
		MarketBot:    loadedConfig.MarketBotEnabled != nil && *loadedConfig.MarketBotEnabled,
		ServerCount:  len(globalRegistry.All()),
	}
}

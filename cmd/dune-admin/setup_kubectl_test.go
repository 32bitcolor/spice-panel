package main

import (
	"strings"
	"testing"
)

func TestSelectBattlegroup(t *testing.T) {
	t.Parallel()

	if got := selectBattlegroup(nil, func(string, string) string { return "" }); got != "" {
		t.Fatalf("expected empty selection for no battlegroups, got %q", got)
	}

	if got := selectBattlegroup([]string{"alpha"}, func(string, string) string { return "1" }); got != "alpha" {
		t.Fatalf("expected single battlegroup selection, got %q", got)
	}

	groups := []string{"alpha", "beta", "gamma"}
	if got := selectBattlegroup(groups, func(string, string) string { return "2" }); got != "beta" {
		t.Fatalf("expected selection index 2 => beta, got %q", got)
	}

	if got := selectBattlegroup(groups, func(string, string) string { return "99" }); got != "alpha" {
		t.Fatalf("expected invalid selection to fall back to first, got %q", got)
	}
}

func TestSSHDefaultHost(t *testing.T) {
	// Serial: subtests mutate SSH_HOST env var.
	tests := []struct {
		name    string
		cfg     appConfig
		envHost string
		want    string
	}{
		{
			name:    "config value takes priority over env",
			cfg:     appConfig{SSHHost: "10.0.0.5:22"},
			envHost: "192.168.0.72:22",
			want:    "10.0.0.5:22",
		},
		{
			name:    "env var used when config is empty",
			cfg:     appConfig{},
			envHost: "192.168.1.100:22",
			want:    "192.168.1.100:22",
		},
		{
			name:    "hardcoded default when both empty",
			cfg:     appConfig{},
			envHost: "",
			want:    "192.168.0.72:22",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SSH_HOST", tt.envHost)
			got := sshDefaultHost(tt.cfg)
			if got != tt.want {
				t.Errorf("sshDefaultHost() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSSHDefaultUser(t *testing.T) {
	// Serial: subtests mutate SSH_USER env var.
	tests := []struct {
		name    string
		cfg     appConfig
		envUser string
		want    string
	}{
		{
			name:    "config value takes priority over env",
			cfg:     appConfig{SSHUser: "admin"},
			envUser: "dune",
			want:    "admin",
		},
		{
			name:    "env var used when config is empty",
			cfg:     appConfig{},
			envUser: "myuser",
			want:    "myuser",
		},
		{
			name:    "hardcoded default when both empty",
			cfg:     appConfig{},
			envUser: "",
			want:    "dune",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SSH_USER", tt.envUser)
			got := sshDefaultUser(tt.cfg)
			if got != tt.want {
				t.Errorf("sshDefaultUser() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSSHKeyNotFoundMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		configuredKey string
		wantContains  string
	}{
		{
			name:          "configured path shown when set",
			configuredKey: "/home/dune-admin/.ssh/id_rsa",
			wantContains:  "/home/dune-admin/.ssh/id_rsa",
		},
		{
			name:          "auto-detect candidates shown when no configured path",
			configuredKey: "",
			wantContains:  "~/.dune-admin/sshKey",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := sshKeyNotFoundMessage(tt.configuredKey)
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("sshKeyNotFoundMessage(%q) = %q, want it to contain %q", tt.configuredKey, got, tt.wantContains)
			}
		})
	}
}

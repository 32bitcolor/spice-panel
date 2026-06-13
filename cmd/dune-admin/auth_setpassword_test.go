package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSetLocalPassword(t *testing.T) {
	t.Run("writes hash and username to existing config", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		if err := os.WriteFile(path, []byte("db_host: 1.2.3.4\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := setLocalPassword(path, "admin", "s3cret"); err != nil {
			t.Fatalf("setLocalPassword: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var cfg appConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.DBHost != "1.2.3.4" {
			t.Errorf("existing config lost: db_host = %q", cfg.DBHost)
		}
		if cfg.AuthLocalUsername != "admin" {
			t.Errorf("username = %q, want admin", cfg.AuthLocalUsername)
		}
		if !checkPassword(cfg.AuthLocalPasswordHash, "s3cret") {
			t.Error("stored hash does not match password")
		}
	})

	t.Run("creates config when missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := setLocalPassword(path, "owner", "pw"); err != nil {
			t.Fatalf("setLocalPassword: %v", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var cfg appConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			t.Fatal(err)
		}
		if !checkPassword(cfg.AuthLocalPasswordHash, "pw") {
			t.Error("stored hash does not match password")
		}
	})

	t.Run("rejects empty password", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := setLocalPassword(path, "admin", ""); err == nil {
			t.Error("empty password accepted")
		}
	})

	t.Run("rejects empty username", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		if err := setLocalPassword(path, "", "pw"); err == nil {
			t.Error("empty username accepted")
		}
	})
}

func TestApplyNewLocalPassword(t *testing.T) {
	t.Run("hashes plaintext into hash field and clears it", func(t *testing.T) {
		cfg := appConfig{AuthLocalPasswordNew: "newpw", AuthLocalPasswordHash: "old"}
		if err := applyNewLocalPassword(&cfg); err != nil {
			t.Fatalf("applyNewLocalPassword: %v", err)
		}
		if cfg.AuthLocalPasswordNew != "" {
			t.Error("plaintext field not cleared")
		}
		if !checkPassword(cfg.AuthLocalPasswordHash, "newpw") {
			t.Error("hash does not match new password")
		}
	})

	t.Run("no-op when plaintext absent", func(t *testing.T) {
		cfg := appConfig{AuthLocalPasswordHash: "keep"}
		if err := applyNewLocalPassword(&cfg); err != nil {
			t.Fatalf("applyNewLocalPassword: %v", err)
		}
		if cfg.AuthLocalPasswordHash != "keep" {
			t.Errorf("hash changed: %q", cfg.AuthLocalPasswordHash)
		}
	})
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// setLocalPassword updates (or creates) the config file at path with the
// local dashboard login credentials. This is the lockout-recovery path: it
// works without a running server or a reachable Discord API.
func setLocalPassword(path, username, plain string) error {
	if username == "" {
		return fmt.Errorf("username must not be empty")
	}
	if plain == "" {
		return fmt.Errorf("password must not be empty")
	}
	var cfg appConfig
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse existing config: %w", err)
		}
	}
	hash, err := hashPassword(plain)
	if err != nil {
		return err
	}
	cfg.AuthLocalUsername = username
	cfg.AuthLocalPasswordHash = hash
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// applyNewLocalPassword hashes the write-only auth_local_password_new field
// into auth_local_password_hash and clears the plaintext. Called by
// handleSaveConfig before persisting, so the plaintext never reaches disk.
func applyNewLocalPassword(cfg *appConfig) error {
	if cfg.AuthLocalPasswordNew == "" {
		return nil
	}
	hash, err := hashPassword(cfg.AuthLocalPasswordNew)
	if err != nil {
		return err
	}
	cfg.AuthLocalPasswordHash = hash
	cfg.AuthLocalPasswordNew = ""
	return nil
}

// runSetPasswordMode implements the --set-password CLI flag: prompts for a
// username and password on stdin, writes them to config.yaml, and exits.
func runSetPasswordMode() error {
	r := bufio.NewReader(os.Stdin)
	fmt.Print("Dashboard username [admin]: ")
	username, _ := r.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}
	fmt.Print("New password: ")
	password, _ := r.ReadString('\n')
	password = strings.TrimSpace(password)
	if err := setLocalPassword(configPath(), username, password); err != nil {
		return err
	}
	fmt.Printf("Local login updated for %q in %s\n", username, configPath())
	fmt.Println("Set auth_enabled: true in the config to enforce dashboard login.")
	return nil
}

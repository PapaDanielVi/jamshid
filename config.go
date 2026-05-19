package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DirEntry tracks a directory linked to a profile.
type DirEntry struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	Profile string `json:"profile"`
}

// Config holds global jamshid settings.
type Config struct {
	Version       string              `json:"version"`
	GlobalProfile string              `json:"global_profile,omitempty"`
	Profiles      map[string]Profile  `json:"profiles,omitempty"`
	VaultRemote   string              `json:"vault_remote,omitempty"`
	LinkedDirs    map[string]DirEntry `json:"linked_dirs,omitempty"`
}

func jamshidDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
	}
	return filepath.Join(home, ".config/jamshid"), nil
}

func configPath() (string, error) {
	dir, err := jamshidDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig reads ~/.config/jamshid/config.json. Creates defaults if missing.
func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Version: "1", Profiles: make(map[string]Profile)}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	if cfg.LinkedDirs == nil {
		cfg.LinkedDirs = make(map[string]DirEntry)
	}
	return &cfg, nil
}

// SaveConfig writes config to ~/.config/jamshid/config.json.
func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

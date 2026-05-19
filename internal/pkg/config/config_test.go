package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PapaDanielVi/jamshid/internal/pkg/models"
)

func TestLoadSaveConfig(t *testing.T) {
	dir := t.TempDir()
	// Override config path for testing
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Version != "1" {
		t.Errorf("Version = %q, want %q", cfg.Version, "1")
	}

	cfg.GlobalProfile = "personal"
	cfg.Profiles["test"] = models.Profile{Name: "test"}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	cfg2, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig second: %v", err)
	}
	if cfg2.GlobalProfile != "personal" {
		t.Errorf("GlobalProfile = %q, want %q", cfg2.GlobalProfile, "personal")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	configPath := filepath.Join(dir, ".config/jamshid/config.json")
	_ = os.MkdirAll(filepath.Dir(configPath), 0755)
	_ = os.WriteFile(configPath, []byte("{invalid json"), 0644)

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

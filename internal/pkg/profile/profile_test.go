package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/models"
)

func TestAddDeleteProfile(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if err := AddProfile(cfg, "test"); err != nil {
		t.Fatalf("AddProfile: %v", err)
	}
	if _, ok := cfg.Profiles["test"]; !ok {
		t.Error("profile not added to config")
	}

	// Check directory created
	profileDir := filepath.Join(dir, ".config/jamshid/profiles/test/.claude")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Error("profile directory not created")
	}

	if err := DeleteProfile(cfg, "test"); err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}
	if _, ok := cfg.Profiles["test"]; ok {
		t.Error("profile not removed from config")
	}
}

func TestGetProfile(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]models.Profile{
			"work": {Name: "work", Model: "claude-opus-4-7"},
		},
	}
	p, ok := GetProfile(cfg, "work")
	if !ok {
		t.Fatal("profile not found")
	}
	if p.Model != "claude-opus-4-7" {
		t.Errorf("Model = %q, want %q", p.Model, "claude-opus-4-7")
	}
}

func TestListProfiles(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]models.Profile{
			"a": {Name: "a"},
			"b": {Name: "b"},
		},
	}
	names := ListProfiles(cfg)
	if len(names) != 2 {
		t.Errorf("len(ListProfiles) = %d, want 2", len(names))
	}
}

func TestLinkUnlinkProfile(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	cfg, _ := config.LoadConfig()
	cfg.LinkedDirs = make(map[string]config.DirEntry)
	if err := AddProfile(cfg, "myprofile"); err != nil {
		t.Fatalf("AddProfile: %v", err)
	}

	cwd := t.TempDir()
	if err := LinkProfile(cfg, cwd, "myprofile", false); err != nil {
		t.Fatalf("LinkProfile: %v", err)
	}

	link := filepath.Join(cwd, ".claude", "settings.local.json")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink: %v", err)
	}
	if filepath.Base(filepath.Dir(filepath.Dir(target))) != "myprofile" {
		t.Errorf("symlink target incorrect: %s", target)
	}

	if err := UnlinkProfile(cfg, cwd); err != nil {
		t.Fatalf("UnlinkProfile: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Error("symlink not removed")
	}
}

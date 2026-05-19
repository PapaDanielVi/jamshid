package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// McpServer defines an MCP server configuration.
type McpServer struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// Profile holds configuration for a Claude Code profile.
type Profile struct {
	Name         string            `json:"name"`
	EnvVars      map[string]string `json:"env_vars,omitempty"`
	ClaudeConfig map[string]any    `json:"claude_config,omitempty"`
	McpServers   []McpServer       `json:"mcp_servers,omitempty"`
	Model        string            `json:"model,omitempty"`
	Timeout      string            `json:"timeout,omitempty"`
}

// ProfileDir returns the directory for a profile.
func ProfileDir(name string) (string, error) {
	dir, err := jamshidDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles", name), nil
}

// AddProfile creates a new profile with the given name.
// Optional importPath can be provided to import settings from an existing file.
func AddProfile(cfg *Config, name string, importPath ...string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if _, exists := cfg.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	cfg.Profiles[name] = Profile{
		Name:    name,
		EnvVars: make(map[string]string),
	}
	// Create profile directory
	dir, err := ProfileDir(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0755); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}

	// Handle optional import path
	if len(importPath) > 0 && importPath[0] != "" {
		src := importPath[0]
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read import file: %w", err)
		}
		dst := filepath.Join(dir, ".claude", filepath.Base(src))
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("write imported settings: %w", err)
		}
	} else {
		// Create default settings.json
		settings := map[string]any{}
		data, err := json.MarshalIndent(settings, "", "    ")
		if err != nil {
			return fmt.Errorf("marshal settings: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".claude/settings.json"), data, 0644); err != nil {
			return fmt.Errorf("write settings: %w", err)
		}
	}
	return nil
}

// DeleteProfile removes a profile by name.
func DeleteProfile(cfg *Config, name string) error {
	if _, exists := cfg.Profiles[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(cfg.Profiles, name)
	// Remove profile directory
	dir, err := ProfileDir(name)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove profile dir: %w", err)
	}
	return nil
}

// GetProfile returns a profile by name.
func GetProfile(cfg *Config, name string) (Profile, bool) {
	p, ok := cfg.Profiles[name]
	return p, ok
}

// ListProfiles returns all profile names.
func ListProfiles(cfg *Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	return names
}

// LinkProfile symlinks a profile's .claude dir to cwd.
func LinkProfile(cfg *Config, cwd, profileName string) error {
	if _, exists := cfg.Profiles[profileName]; !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}

	hash := DirHash(cwd)
	dir, err := ProfileDir(profileName)
	if err != nil {
		return err
	}

	claudeTarget := filepath.Join(dir, ".claude")
	claudeLink := filepath.Join(cwd, ".claude")

	// If .claude exists as a real directory, back it up
	if info, err := os.Lstat(claudeLink); err == nil {
		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			backup := claudeLink + ".bak"
			if err := os.Rename(claudeLink, backup); err != nil {
				return fmt.Errorf("backup .claude: %w", err)
			}
		} else if info.Mode()&os.ModeSymlink != 0 {
			_ = os.Remove(claudeLink)
		}
	}

	if err := os.Symlink(claudeTarget, claudeLink); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}

	cfg.LinkedDirs[hash] = DirEntry{Path: cwd, Hash: hash, Profile: profileName}
	return nil
}

// UnlinkProfile removes the .claude symlink from cwd.
func UnlinkProfile(cfg *Config, cwd string) error {
	hash := DirHash(cwd)
	delete(cfg.LinkedDirs, hash)

	claudeLink := filepath.Join(cwd, ".claude")
	if info, err := os.Lstat(claudeLink); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(claudeLink); err != nil {
			return fmt.Errorf("remove symlink: %w", err)
		}
	}
	return nil
}

// IsGitRepo checks if the directory is inside a git repo.
func IsGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetActiveProfile returns the active profile for cwd.
func GetActiveProfile(cfg *Config, cwd string) string {
	hash := DirHash(cwd)
	if entry, ok := cfg.LinkedDirs[hash]; ok {
		return entry.Profile
	}
	return cfg.GlobalProfile
}

package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/hash"
	"github.com/PapaDanielVi/jamshid/internal/pkg/models"
)

// ProfileDir returns the directory for a profile.
func ProfileDir(name string) (string, error) {
	dir, err := config.JamshidDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles", name), nil
}

// AddProfile creates a new profile with the given name.
// Optional importPath can be provided to import settings from an existing file or directory.
func AddProfile(cfg *config.Config, name string, importPath ...string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if _, exists := cfg.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	cfg.Profiles[name] = models.Profile{
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
		// Check if src is a directory (.claude folder)
		if info, err := os.Stat(src); err == nil && info.IsDir() {
			// Copy entire directory
			if err := copyDir(src, filepath.Join(dir, ".claude")); err != nil {
				return fmt.Errorf("copy .claude dir: %w", err)
			}
		} else {
			// Copy single file
			data, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("read import file: %w", err)
			}
			dst := filepath.Join(dir, ".claude", filepath.Base(src))
			if err := os.WriteFile(dst, data, 0644); err != nil {
				return fmt.Errorf("write imported settings: %w", err)
			}
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

// copyDir copies an entire directory recursively.
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteProfile removes a profile by name.
func DeleteProfile(cfg *config.Config, name string) error {
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
func GetProfile(cfg *config.Config, name string) (models.Profile, bool) {
	p, ok := cfg.Profiles[name]
	return p, ok
}

// ListProfiles returns all profile names.
func ListProfiles(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	return names
}

// LinkProfile symlinks a profile's .claude dir to cwd.
func LinkProfile(cfg *config.Config, cwd, profileName string, force bool) error {
	if _, exists := cfg.Profiles[profileName]; !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}

	// Check if .claude/settings.local.json exists and force is not set
	settingsLocal := filepath.Join(cwd, ".claude", "settings.local.json")
	if _, err := os.Stat(settingsLocal); err == nil && !force {
		return fmt.Errorf(".claude/settings.local.json already exists, use --force to overwrite")
	}

	hash := hash.DirHash(cwd)
	dir, err := ProfileDir(profileName)
	if err != nil {
		return err
	}

	claudeTarget := filepath.Join(dir, ".claude")
	claudeLink := filepath.Join(cwd, ".claude")

	// If .claude exists as a real directory, handle it
	if info, err := os.Lstat(claudeLink); err == nil {
		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			if force {
				// If force, remove the existing directory
				if err := os.RemoveAll(claudeLink); err != nil {
					return fmt.Errorf("remove .claude: %w", err)
				}
			} else {
				backup := claudeLink + ".bak"
				if err := os.Rename(claudeLink, backup); err != nil {
					return fmt.Errorf("backup .claude: %w", err)
				}
			}
		} else if info.Mode()&os.ModeSymlink != 0 {
			_ = os.Remove(claudeLink)
		}
	}

	if err := os.Symlink(claudeTarget, claudeLink); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}

	cfg.LinkedDirs[hash] = config.DirEntry{Path: cwd, Hash: hash, Profile: profileName}
	return nil
}

// UnlinkProfile removes the .claude symlink from cwd.
func UnlinkProfile(cfg *config.Config, cwd string) error {
	hash := hash.DirHash(cwd)
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
func GetActiveProfile(cfg *config.Config, cwd string) string {
	hash := hash.DirHash(cwd)
	if entry, ok := cfg.LinkedDirs[hash]; ok {
		return entry.Profile
	}
	return cfg.GlobalProfile
}

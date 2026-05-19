package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/constants"
	"github.com/PapaDanielVi/jamshid/internal/pkg/hash"
	"github.com/PapaDanielVi/jamshid/internal/pkg/models"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
	ErrProfileExists   = errors.New("profile already exists")
	ErrEmptyName       = errors.New("profile name cannot be empty")
	ErrSettingsExists  = errors.New(".claude/settings.local.json already exists, use --force to overwrite")
)

// mcpConfigFiles lists known MCP config file names that may appear
// inside a .claude directory or project root.
var mcpConfigFiles = []string{
	".mcp.json",
	"mcp.json",
	"mcp_servers.json",
}

// ProfileDir returns the directory path for a profile.
func ProfileDir(name string) (string, error) {
	dir, err := config.JamshidDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, constants.DirProfiles, name), nil
}

func AddProfile(cfg *config.Config, name string, importPath ...string) error {
	if name == "" {
		return ErrEmptyName
	}
	if _, exists := cfg.Profiles[name]; exists {
		return fmt.Errorf("profile %q: %w", name, ErrProfileExists)
	}
	cfg.Profiles[name] = models.Profile{
		Name:    name,
		EnvVars: make(map[string]string),
	}
	dir, err := ProfileDir(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, constants.DirClaude), constants.DefaultDirPerm); err != nil {
		return fmt.Errorf("create profile dir: %w", err)
	}

	if len(importPath) > 0 && importPath[0] != "" {
		src := importPath[0]
		if info, err := os.Stat(src); err == nil && info.IsDir() {
			if err := copyDir(src, filepath.Join(dir, constants.DirClaude)); err != nil {
				return fmt.Errorf("copy .claude dir: %w", err)
			}
			// Also copy MCP config files from the project root (parent of .claude dir)
			srcDir := filepath.Dir(src)
			copyMcpConfigs(srcDir, dir)
		} else {
			data, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("read import file: %w", err)
			}
			dst := filepath.Join(dir, constants.DirClaude, filepath.Base(src))
			if err := os.WriteFile(dst, data, constants.DefaultFilePerm); err != nil {
				return fmt.Errorf("write imported settings: %w", err)
			}
		}
	} else {
		settings := map[string]any{}
		data, err := json.MarshalIndent(settings, "", "    ")
		if err != nil {
			return fmt.Errorf("marshal settings: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, constants.DirClaude, constants.FileSettingsJSON), data, constants.DefaultFilePerm); err != nil {
			return fmt.Errorf("write settings: %w", err)
		}
	}
	return nil
}

// copyMcpConfigs copies known MCP config files from srcDir to dstDir.
func copyMcpConfigs(srcDir, dstDir string) {
	for _, name := range mcpConfigFiles {
		srcFile := filepath.Join(srcDir, name)
		if _, err := os.Stat(srcFile); err == nil {
			data, err := os.ReadFile(srcFile)
			if err != nil {
				continue
			}
			_ = os.WriteFile(filepath.Join(dstDir, name), data, constants.DefaultFilePerm)
		}
	}
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, constants.DefaultDirPerm); err != nil {
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
			info, err := os.Stat(srcPath)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, info.Mode().Perm()); err != nil {
				return err
			}
		}
	}
	return nil
}

func DeleteProfile(cfg *config.Config, name string) error {
	if _, exists := cfg.Profiles[name]; !exists {
		return fmt.Errorf("profile %q: %w", name, ErrProfileNotFound)
	}
	delete(cfg.Profiles, name)
	dir, err := ProfileDir(name)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove profile dir: %w", err)
	}
	return nil
}

func GetProfile(cfg *config.Config, name string) (models.Profile, bool) {
	p, ok := cfg.Profiles[name]
	return p, ok
}

func ListProfiles(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// mcpSymlinkTargets returns the MCP config symlink targets for a profile.
// Returns a map of link path (in cwd) -> target path (in profile dir).
func mcpSymlinkTargets(cwd, profileDir string) map[string]string {
	targets := make(map[string]string)
	for _, name := range mcpConfigFiles {
		target := filepath.Join(profileDir, name)
		if _, err := os.Stat(target); err == nil {
			targets[filepath.Join(cwd, name)] = target
		}
	}
	return targets
}

func LinkProfile(cfg *config.Config, cwd, profileName string, force bool) error {
	if _, exists := cfg.Profiles[profileName]; !exists {
		return fmt.Errorf("profile %q: %w", profileName, ErrProfileNotFound)
	}

	settingsLocal := filepath.Join(cwd, constants.DirClaude, constants.FileSettingsLocal)
	if _, err := os.Stat(settingsLocal); err == nil && !force {
		return ErrSettingsExists
	}

	hash := hash.DirHash(cwd)
	dir, err := ProfileDir(profileName)
	if err != nil {
		return err
	}

	claudeTarget := filepath.Join(dir, constants.DirClaude, constants.FileSettingsLocal)
	claudeLink := filepath.Join(cwd, constants.DirClaude, constants.FileSettingsLocal)

	if info, err := os.Lstat(claudeLink); err == nil {
		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			if force {
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

	if err := os.MkdirAll(filepath.Join(cwd, constants.DirClaude), constants.DefaultDirPerm); err != nil {
		return fmt.Errorf("create .claude dir: %w", err)
	}
	if err := os.Symlink(claudeTarget, claudeLink); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}

	// Symlink MCP config files
	for linkPath, targetPath := range mcpSymlinkTargets(cwd, dir) {
		_ = os.Remove(linkPath)
		_ = os.Symlink(targetPath, linkPath)
	}

	cfg.LinkedDirs[hash] = config.DirEntry{Path: cwd, Hash: hash, Profile: profileName}
	return nil
}

func UnlinkProfile(cfg *config.Config, cwd string) error {
	hash := hash.DirHash(cwd)
	entry, hasEntry := cfg.LinkedDirs[hash]
	delete(cfg.LinkedDirs, hash)

	claudeLink := filepath.Join(cwd, constants.DirClaude, constants.FileSettingsLocal)
	if info, err := os.Lstat(claudeLink); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(claudeLink); err != nil {
			return fmt.Errorf("remove symlink: %w", err)
		}
	}

	// Remove MCP config symlinks
	if hasEntry {
		dir, err := ProfileDir(entry.Profile)
		if err == nil {
			for linkPath := range mcpSymlinkTargets(cwd, dir) {
				if info, err := os.Lstat(linkPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
					_ = os.Remove(linkPath)
				}
			}
		}
	}

	return nil
}

func GetActiveProfile(cfg *config.Config, cwd string) string {
	hash := hash.DirHash(cwd)
	if entry, ok := cfg.LinkedDirs[hash]; ok {
		return entry.Profile
	}
	return ""
}

// ProfilePath returns the filesystem path to a profile's config directory.
func ProfilePath(name string) (string, error) {
	dir, err := ProfileDir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, constants.DirClaude), nil
}


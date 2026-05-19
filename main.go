package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		// Launch TUI
		cfg, err := LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if cfg.LinkedDirs == nil {
			cfg.LinkedDirs = make(map[string]DirEntry)
		}
		cwd, _ := os.Getwd()

		// Auto-detect existing Claude settings not in our config
		if IsGitRepo(cwd) && !isLinked(cwd, cfg) {
			if hasClaudeSettings(cwd) {
				fmt.Println("Found existing Claude settings in this repo.")
				fmt.Println("Would you like to import them as a jamshid profile? (not yet implemented)")
				// TODO: Implement interactive profile creation from existing settings
			}
		}

		m := newTUI(cfg, cwd)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if cfg.LinkedDirs == nil {
		cfg.LinkedDirs = make(map[string]DirEntry)
	}

	switch os.Args[1] {
	case "add":
		cmdAdd(cfg, os.Args[2:])
	case "delete":
		cmdDelete(cfg, os.Args[2:])
	case "list":
		cmdList(cfg)
	case "local":
		cmdLocal(cfg, os.Args[2:])
	case "unlink":
		cmdUnlink(cfg)
	case "global":
		cmdGlobal(cfg, os.Args[2:])
	case "vault":
		cmdVault(cfg, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func cmdAdd(cfg *Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid add <name>")
		os.Exit(1)
	}
	name := args[0]
	if err := AddProfile(cfg, name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Profile %q created\n", name)
}

func cmdDelete(cfg *Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid delete <name>")
		os.Exit(1)
	}
	name := args[0]
	if err := DeleteProfile(cfg, name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Profile %q deleted\n", name)
}

func cmdList(cfg *Config) {
	cwd, _ := os.Getwd()
	active := GetActiveProfile(cfg, cwd)

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured")
		return
	}
	for name := range cfg.Profiles {
		marker := "  "
		if name == active {
			marker = "* "
		}
		if name == cfg.GlobalProfile {
			fmt.Printf("%s%s (global)\n", marker, name)
		} else {
			fmt.Printf("%s%s\n", marker, name)
		}
	}
}

func cmdLocal(cfg *Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid local <profile>")
		os.Exit(1)
	}
	profileName := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !IsGitRepo(cwd) {
		fmt.Fprintln(os.Stderr, "Error: not a git repository")
		os.Exit(1)
	}
	if err := LinkProfile(cfg, cwd, profileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := EnsureGitignore(cwd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Linked profile %q to %s\n", profileName, cwd)
}

func cmdUnlink(cfg *Config) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := UnlinkProfile(cfg, cwd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Unlinked profile from %s\n", cwd)
}

func cmdGlobal(cfg *Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid global <profile>")
		os.Exit(1)
	}
	name := args[0]
	if _, ok := cfg.Profiles[name]; !ok {
		fmt.Fprintf(os.Stderr, "Error: profile %q not found\n", name)
		os.Exit(1)
	}
	cfg.GlobalProfile = name
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Global profile set to %q\n", name)
}

func cmdVault(cfg *Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid vault <init|sync>")
		os.Exit(1)
	}

	// Check gh CLI before any vault operation
	if err := checkGhAuth(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: jamshid vault init <url>")
			os.Exit(1)
		}
		cfg.VaultRemote = args[1]
		if err := SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Vault remote set to %s\n", args[1])
	case "sync":
		fmt.Println("Vault sync not yet implemented")
	default:
		fmt.Fprintf(os.Stderr, "Unknown vault command: %s\n", args[0])
		os.Exit(1)
	}
}

// isLinked checks if cwd is already linked in config.
func isLinked(cwd string, cfg *Config) bool {
	hash := DirHash(cwd)
	_, linked := cfg.LinkedDirs[hash]
	return linked
}

// hasClaudeSettings checks if .claude/ directory exists with settings.
func hasClaudeSettings(cwd string) bool {
	claudeDir := filepath.Join(cwd, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	if info, err := os.Stat(claudeDir); err == nil && info.IsDir() {
		if _, err := os.Stat(settingsPath); err == nil {
			return true
		}
	}
	return false
}

func init() {
	// Ensure config directory exists on startup
	dir, err := jamshidDir()
	if err == nil {
		_ = os.MkdirAll(filepath.Join(dir, "profiles"), 0755)
	}
}

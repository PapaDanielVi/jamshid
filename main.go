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
	case "link":
		cmdLink(cfg, os.Args[2:])
	case "unlink":
		cmdUnlink(cfg)
	case "global":
		cmdGlobal(cfg, os.Args[2:])
	case "vault":
		cmdVault(cfg, os.Args[2:])
	case "help", "--help", "-h":
		cmdHelp()
	case "completion":
		cmdCompletion(os.Args[2:])
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
	cwd, _ := os.Getwd()

	var importPath string
	settingsLocal := filepath.Join(cwd, ".claude", "settings.local.json")

	// Task 2: Check for existing settings.local.json in cwd
	if _, err := os.Stat(settingsLocal); err == nil {
		if !isLinked(cwd, cfg) {
			fmt.Printf("Found existing %s. Create profile from this? (y/n): ", settingsLocal)
			var answer string
			fmt.Scanln(&answer)
			if answer == "y" || answer == "Y" {
				importPath = settingsLocal
			}
		}
	}

	// Task 3: If no settings.local.json, ask user for file path
	if importPath == "" {
		if _, err := os.Stat(settingsLocal); os.IsNotExist(err) {
			fmt.Print("No .claude/settings.local.json found. Provide a file path to import (leave empty to skip): ")
			var pathInput string
			fmt.Scanln(&pathInput)
			if pathInput != "" {
				if _, err := os.Stat(pathInput); err == nil {
					importPath = pathInput
				} else {
					fmt.Fprintf(os.Stderr, "Warning: file %q not found, skipping import\n", pathInput)
				}
			}
		}
	}

	if err := AddProfile(cfg, name, importPath); err != nil {
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

// linkProfileToCwd links a profile to the current working directory.
func linkProfileToCwd(cfg *Config, profileName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	if !IsGitRepo(cwd) {
		return fmt.Errorf("not a git repository")
	}
	if err := LinkProfile(cfg, cwd, profileName); err != nil {
		return err
	}
	if err := EnsureGitignore(cwd); err != nil {
		return err
	}
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Linked profile %q to %s\n", profileName, cwd)
	return nil
}

// cmdLink links a profile to current directory, with interactive selection if no profile given.
func cmdLink(cfg *Config, args []string) {
	var profileName string

	if len(args) < 1 {
		// Interactive mode: list profiles and prompt for selection
		profiles := ListProfiles(cfg)
		if len(profiles) == 0 {
			fmt.Println("No profiles available. Create one with 'jamshid add <name>'")
			os.Exit(1)
		}
		fmt.Println("Available profiles:")
		for i, p := range profiles {
			fmt.Printf("  %d: %s\n", i+1, p)
		}
		fmt.Print("Select profile (number or name): ")
		var input string
		fmt.Scanln(&input)

		// Check if input is a number
		profileName = input
		if _, ok := cfg.Profiles[profileName]; !ok {
			// Try to parse as number
			found := false
			for i, p := range profiles {
				if input == fmt.Sprintf("%d", i+1) {
					profileName = p
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "Error: profile %q not found\n", input)
				os.Exit(1)
			}
		}
	} else {
		profileName = args[0]
	}

	if _, ok := cfg.Profiles[profileName]; !ok {
		fmt.Fprintf(os.Stderr, "Error: profile %q not found\n", profileName)
		os.Exit(1)
	}

	if err := linkProfileToCwd(cfg, profileName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// cmdHelp displays usage information.
func cmdHelp() {
	fmt.Println("jamshid - Claude Code profile manager")
	fmt.Println("\nUsage: jamshid <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  add <name>          Create a new profile")
	fmt.Println("  delete <name>       Delete a profile")
	fmt.Println("  list                List all profiles")
	fmt.Println("  link [profile]      Link profile to current directory (interactive if no profile given)")
	fmt.Println("  unlink              Unlink profile from current directory")
	fmt.Println("  global <profile>    Set global profile")
	fmt.Println("  vault <init|sync>   Manage Git vault")
	fmt.Println("  help, --help, -h    Show this help message")
	fmt.Println("  completion bash     Generate bash completion script")
}

// cmdCompletion generates shell completion scripts.
func cmdCompletion(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid completion bash")
		os.Exit(1)
	}
	switch args[0] {
	case "bash":
		generateBashCompletion()
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s\n", args[0])
		os.Exit(1)
	}
}

// generateBashCompletion outputs bash completion script.
func generateBashCompletion() {
	script := `# bash completion for jamshid
_jamshid() {
    local cur prev words cword
    _init_completion || return

    local commands="add delete list link unlink global vault help completion"
    local profiles=$(jamshid list 2>/dev/null | sed 's/^[* ] //')

    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "$commands" -- "$cur"))
        return
    fi

    case ${words[1]} in
        link|global|delete)
            COMPREPLY=($(compgen -W "$profiles" -- "$cur"))
            ;;
        vault)
            COMPREPLY=($(compgen -W "init sync" -- "$cur"))
            ;;
    esac
}
complete -F _jamshid jamshid
`
	fmt.Print(script)
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

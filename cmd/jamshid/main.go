package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/constants"
	"github.com/PapaDanielVi/jamshid/internal/pkg/gitvault"
	"github.com/PapaDanielVi/jamshid/internal/pkg/profile"
	"github.com/PapaDanielVi/jamshid/internal/tui"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if cfg.LinkedDirs == nil {
		cfg.LinkedDirs = make(map[string]config.DirEntry)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: get working directory: %v\n", err)
		os.Exit(1)
	}

	checkLinkedDir(cfg, cwd)

	if len(os.Args) < 2 {
		if len(cfg.Profiles) == 0 && cfg.VaultRemote == "" {
			promptFirstRun(cfg)
		}
		m := tui.NewTUI(cfg, cwd)
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		selectedCmd, selectedArg, selectedSubcmd := m.SelectedCommand()
		if selectedCmd != "" {
			args := []string{}
			if selectedSubcmd != "" {
				args = append(args, selectedSubcmd)
			}
			if selectedArg != "" {
				args = append(args, selectedArg)
			}
			executeCommand(cfg, selectedCmd, args, cwd)
		}
		os.Exit(0)
	}

	executeCommand(cfg, os.Args[1], os.Args[2:], cwd)
}

func checkLinkedDir(cfg *config.Config, cwd string) {
	name := profile.GetActiveProfile(cfg, cwd)
	if name != "" {
		fmt.Printf("Linked to profile: %s\n", name)
	}
}

func executeCommand(cfg *config.Config, cmd string, args []string, cwd string) {
	switch cmd {
	case "add":
		cmdAdd(cfg, args, cwd)
	case "delete":
		cmdDelete(cfg, args)
	case "list":
		cmdList(cfg)
	case "link":
		cmdLink(cfg, args, cwd)
	case "unlink":
		cmdUnlink(cfg, cwd)
	case "env":
		cmdEnv(cfg, args)
	case "status":
		cmdStatus(cfg, cwd)
	case "vault":
		cmdVault(cfg, args)
	case "version", "--version", "-v":
		fmt.Printf("jamshid %s\n", Version)
	case "help", "--help", "-h":
		cmdHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func cmdAdd(cfg *config.Config, args []string, cwd string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid add <name>")
		os.Exit(1)
	}
	name := args[0]

	var importPath string
	settingsLocal := filepath.Join(cwd, constants.DirClaude, constants.FileSettingsLocal)

	if _, err := os.Stat(settingsLocal); err == nil {
		if !isLinked(cwd, cfg) {
			fmt.Printf("Found existing %s. Create profile from this? (y/n): ", settingsLocal)
			var answer string
			_, _ = fmt.Scanln(&answer)
			if answer == "y" || answer == "Y" {
				importPath = filepath.Join(cwd, constants.DirClaude)
			}
		}
	}

	if importPath == "" {
		if _, err := os.Stat(settingsLocal); os.IsNotExist(err) {
			fmt.Print("No .claude/settings.local.json found. Provide a file path to import (leave empty to skip): ")
			var pathInput string
			_, _ = fmt.Scanln(&pathInput)
			if pathInput != "" {
				if _, err := os.Stat(pathInput); err == nil {
					importPath = pathInput
				} else {
					fmt.Fprintf(os.Stderr, "Warning: file %q not found, skipping import\n", pathInput)
				}
			}
		}
	}

	if err := profile.AddProfile(cfg, name, importPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Profile %q created\n", name)
}

func cmdDelete(cfg *config.Config, args []string) {
	var name string
	if len(args) < 1 {
		profiles := profile.ListProfiles(cfg)
		if len(profiles) == 0 {
			fmt.Println("No profiles available.")
			os.Exit(1)
		}
		fmt.Println("Available profiles:")
		for i, p := range profiles {
			fmt.Printf("  %d: %s\n", i+1, p)
		}
		fmt.Print("Select profile to delete (number or name): ")
		var input string
		_, _ = fmt.Scanln(&input)

		name = input
		if _, ok := cfg.Profiles[name]; !ok {
			found := false
			for i, p := range profiles {
				if input == strconv.Itoa(i+1) {
					name = p
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
		name = args[0]
	}

	if err := profile.DeleteProfile(cfg, name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Profile %q deleted\n", name)
}

func cmdList(cfg *config.Config) {
	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured")
		return
	}
	for _, name := range profile.ListProfiles(cfg) {
		dir, err := profile.ProfilePath(name)
		if err != nil {
			dir = "(unknown path)"
		}
		fmt.Printf("  %s  (%s)\n", name, dir)
	}
}

func cmdLink(cfg *config.Config, args []string, cwd string) {
	var profileName string
	force := false

	args, flags := parseFlags(args)
	for _, f := range flags {
		if f == "force" {
			force = true
		}
	}

	if len(args) < 1 {
		profiles := profile.ListProfiles(cfg)
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
		_, _ = fmt.Scanln(&input)

		profileName = input
		if _, ok := cfg.Profiles[profileName]; !ok {
			found := false
			for i, p := range profiles {
				if input == strconv.Itoa(i+1) {
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

	if err := profile.LinkProfile(cfg, cwd, profileName, force); err != nil {
		if errors.Is(err, profile.ErrSettingsExists) && !force {
			fmt.Print("Remove existing .claude/settings.local.json and link profile? (y/n): ")
			var answer string
			_, _ = fmt.Scanln(&answer)
			if answer == "y" || answer == "Y" {
				if err := profile.LinkProfile(cfg, cwd, profileName, true); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Fprintln(os.Stderr, "Aborted.")
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Linked profile %q to %s\n", profileName, cwd)
}

func parseFlags(args []string) ([]string, []string) {
	var flags []string
	var remaining []string
	for _, arg := range args {
		if len(arg) > 2 && arg[:2] == "--" {
			flags = append(flags, arg[2:])
		} else {
			remaining = append(remaining, arg)
		}
	}
	return remaining, flags
}

func cmdUnlink(cfg *config.Config, cwd string) {
	if err := profile.UnlinkProfile(cfg, cwd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Unlinked profile from %s\n", cwd)
}

func cmdEnv(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid env <profile>")
		os.Exit(1)
	}

	name := args[0]
	if _, ok := cfg.Profiles[name]; !ok {
		fmt.Fprintf(os.Stderr, "Error: profile %q not found\n", name)
		os.Exit(1)
	}

	path, err := profile.ProfilePath(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("export CLAUDE_CONFIG_DIR=%s\n", path)
}

func cmdStatus(cfg *config.Config, cwd string) {
	name := profile.GetActiveProfile(cfg, cwd)
	if name == "" {
		fmt.Printf("No profile linked to %s\n", cwd)
		return
	}
	fmt.Printf("Profile: %s\n", name)
	checkSymlinkStatus(filepath.Join(cwd, constants.DirClaude, constants.FileSettingsLocal),
		constants.DirClaude+"/"+constants.FileSettingsLocal)
	for _, f := range []string{".mcp.json", "mcp.json", "mcp_servers.json"} {
		checkSymlinkStatus(filepath.Join(cwd, f), f)
	}
}

func checkSymlinkStatus(path, label string) {
	info, err := os.Lstat(path)
	if err != nil {
		return
	}
	if info.Mode()&os.ModeSymlink == 0 {
		fmt.Printf("  %s: file (not a symlink)\n", label)
		return
	}
	target, err := os.Readlink(path)
	if err != nil {
		fmt.Printf("  %s: unreadable symlink\n", label)
		return
	}
	if _, err := os.Stat(target); err != nil {
		fmt.Printf("  %s: broken -> %s\n", label, target)
	} else {
		fmt.Printf("  %s: OK -> %s\n", label, target)
	}
}

func cmdVault(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jamshid vault <init|push|pull|status>")
		os.Exit(1)
	}

	ctx := context.Background()

	if err := gitvault.CheckGhAuth(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: jamshid vault init <url>")
			os.Exit(1)
		}
		if err := gitvault.InitVault(ctx, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		cfg.VaultRemote = args[1]
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Vault remote set to %s\n", args[1])
	case "push", "sync":
		if cfg.VaultRemote == "" {
			fmt.Fprintln(os.Stderr, "Error: vault not initialized. Run: jamshid vault init <url>")
			os.Exit(1)
		}
		if err := gitvault.SyncPush(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Vault pushed successfully")
	case "pull":
		if cfg.VaultRemote == "" {
			fmt.Fprintln(os.Stderr, "Error: vault not initialized. Run: jamshid vault init <url>")
			os.Exit(1)
		}
		if err := gitvault.SyncPull(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Vault pulled successfully")
	case "status":
		if cfg.VaultRemote == "" {
			fmt.Fprintln(os.Stderr, "Error: vault not initialized. Run: jamshid vault init <url>")
			os.Exit(1)
		}
		s, err := gitvault.GetVaultStatus(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Remote: %s\n", s.Remote)
		fmt.Printf("Branch: %s\n", s.Branch)
		fmt.Printf("Ahead:  %d\n", s.Ahead)
		fmt.Printf("Behind: %d\n", s.Behind)
	default:
		fmt.Fprintf(os.Stderr, "Unknown vault command: %s\n", args[0])
		os.Exit(1)
	}
}

func cmdHelp() {
	fmt.Println("jamshid - Claude Code profile manager")
	fmt.Println("\nUsage: jamshid <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  add <name>               Create a new profile (imports settings + MCP configs if found)")
	fmt.Println("  delete <name>            Delete a profile")
	fmt.Println("  list                     List all profiles")
	fmt.Println("  link [profile]           Link profile to cwd via symlinks (interactive if no profile given)")
	fmt.Println("  unlink                   Unlink profile symlinks from cwd")
	fmt.Println("  env <profile>            Set CLAUDE_CONFIG_DIR to the profile's config directory")
	fmt.Println("  status                   Show active profile and symlink health for cwd")
	fmt.Println("  vault <init|push|pull|status>  Manage Git vault")
	fmt.Println("  version                  Print version")
	fmt.Println("  help                     Show this help message")
	fmt.Println("\nEnv usage:")
	fmt.Println("  eval $(jamshid env personal)   # Set CLAUDE_CONFIG_DIR in current shell")
	fmt.Println("  eval $(jamshid env work)       # Switch to work profile")
}

func promptFirstRun(cfg *config.Config) {
	fmt.Println("No profiles configured. A Git vault lets you back up and sync profiles.")
	fmt.Println("  1) Initialize a new vault")
	fmt.Println("  2) Skip (create profiles manually)")
	fmt.Print("Choose an option (1/2): ")
	var answer string
	_, _ = fmt.Scanln(&answer)
	if answer != "1" {
		return
	}

	ctx := context.Background()

	if err := gitvault.CheckGhAuth(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("You can set up a vault later with: jamshid vault init <url>")
		return
	}

	fmt.Print("Enter git remote URL for the vault: ")
	var url string
	_, _ = fmt.Scanln(&url)
	if url == "" {
		fmt.Fprintln(os.Stderr, "No URL provided. Skipping vault setup.")
		return
	}

	if err := gitvault.InitVault(ctx, url); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("You can set up a vault later with: jamshid vault init <url>")
		return
	}

	cfg.VaultRemote = url
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Println("Vault initialized.")
}

func isLinked(cwd string, cfg *config.Config) bool {
	return profile.GetActiveProfile(cfg, cwd) != ""
}

func init() {
	dir, err := config.JamshidDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: get jamshid dir: %v\n", err)
		return
	}
	if err := os.MkdirAll(filepath.Join(dir, constants.DirProfiles), constants.DefaultDirPerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error: create profiles dir: %v\n", err)
	}
}

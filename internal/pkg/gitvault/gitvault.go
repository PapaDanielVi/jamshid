package gitvault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/constants"
)

const gitOpTimeout = 60 * time.Second

// vaultCredentialExclusions lists patterns that must never be committed to the vault.
var vaultCredentialExclusions = []string{
	".credentials.json",
	"projects/",
	"history/",
	"todos/",
	"statsig/",
}

// VaultStatus holds information about the vault git repository state.
type VaultStatus struct {
	Remote string
	Branch string
	Ahead  int
	Behind int
}

// CheckGhAuth verifies that the gh CLI is installed and authenticated.
func CheckGhAuth(ctx context.Context) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI is not installed - install it from https://cli.github.com")
	}
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("`gh` is not authenticated - run `gh auth login` first")
	}
	return nil
}

// InitVault initialises a git repository in the jamshid config directory and
// adds the given remote URL as origin.
func InitVault(ctx context.Context, remote string) error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, constants.DefaultDirPerm); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := runGit(ctx, dir, "init"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	if err := runGit(ctx, dir, "remote", "add", "origin", remote); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}
	return nil
}

// EnsureVaultGitignore writes a .gitignore in the vault dir that excludes
// credentials, history, todos, and runtime state. It is idempotent.
func EnsureVaultGitignore(ctx context.Context) error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, constants.DefaultDirPerm); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}
	return ensureGitignoreEntries(filepath.Join(dir, ".gitignore"), vaultCredentialExclusions)
}

func ensureGitignoreEntries(path string, entries []string) error {
	existing := map[string]bool{}
	if data, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			existing[strings.TrimSpace(line)] = true
		}
	}

	var toAdd []string
	for _, e := range entries {
		if !existing[e] {
			toAdd = append(toAdd, e)
		}
	}
	if len(toAdd) == 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constants.DefaultFilePerm)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer func() { _ = f.Close() }()
	for _, e := range toAdd {
		if _, err := fmt.Fprintln(f, e); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
	}
	return nil
}

// CheckPrivateRepo verifies via gh that the given remote URL points to a
// private repository.
func CheckPrivateRepo(ctx context.Context, remote string) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not installed: %w", err)
	}
	cmd := exec.CommandContext(ctx, "gh", "repo", "view", remote, "--json", "isPrivate")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("check repo visibility: %w", err)
	}
	var result struct {
		IsPrivate bool `json:"isPrivate"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return fmt.Errorf("parse gh output: %w", err)
	}
	if !result.IsPrivate {
		return fmt.Errorf("refusing to sync to public repository %q — use a private repo to protect your settings", remote)
	}
	return nil
}

// GetVaultStatus returns the remote URL, current branch, and ahead/behind counts.
func GetVaultStatus(ctx context.Context) (*VaultStatus, error) {
	dir, err := config.JamshidDir()
	if err != nil {
		return nil, err
	}
	remote, err := runGitOutput(ctx, dir, "remote", "get-url", "origin")
	if err != nil {
		return nil, fmt.Errorf("get remote URL: %w", err)
	}
	branch, err := runGitOutput(ctx, dir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("get branch: %w", err)
	}
	// Fetch quietly to update remote-tracking refs; ignore error (network may be unavailable).
	_ = runGit(ctx, dir, "fetch", "--quiet")
	ahead, _ := runGitOutputInt(ctx, dir, "rev-list", "--count", "origin/"+branch+"..HEAD")
	behind, _ := runGitOutputInt(ctx, dir, "rev-list", "--count", "HEAD..origin/"+branch)
	return &VaultStatus{
		Remote: remote,
		Branch: branch,
		Ahead:  ahead,
		Behind: behind,
	}, nil
}

// SyncPush stages all changes, commits if needed, and pushes to the remote.
// On the first push (empty remote), it verifies the remote is private via gh.
func SyncPush(ctx context.Context) error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}
	if err := EnsureVaultGitignore(ctx); err != nil {
		return fmt.Errorf("ensure vault .gitignore: %w", err)
	}
	if isRemoteEmpty(ctx, dir) {
		remote, err := runGitOutput(ctx, dir, "remote", "get-url", "origin")
		if err != nil {
			return fmt.Errorf("get remote URL: %w", err)
		}
		// Only check visibility for GitHub URLs; skip for local paths.
		if isGitHubURL(remote) {
			if err := CheckPrivateRepo(ctx, remote); err != nil {
				return err
			}
		}
	}
	if err := runGit(ctx, dir, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	// diff --cached --quiet exits 1 when there are staged changes.
	if err := runGit(ctx, dir, "diff", "--cached", "--quiet"); err != nil {
		_ = runGit(ctx, dir, "commit", "-m", constants.DefaultCommitMessage)
	}
	if err := runGit(ctx, dir, "push", "origin", "main"); err != nil {
		if err2 := runGit(ctx, dir, "push", "origin", "master"); err2 != nil {
			return fmt.Errorf("git push: %w", err)
		}
	}
	return nil
}

// SyncPull pulls the latest changes from the remote vault.
func SyncPull(ctx context.Context) error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}
	if err := runGit(ctx, dir, "pull", "origin", "main"); err != nil {
		if err2 := runGit(ctx, dir, "pull", "origin", "master"); err2 != nil {
			return fmt.Errorf("git pull: %w", err)
		}
	}
	return nil
}

// isRemoteEmpty reports whether the remote origin has no commits yet.
func isRemoteEmpty(ctx context.Context, dir string) bool {
	out, err := runGitOutput(ctx, dir, "ls-remote", "origin", "HEAD")
	return err == nil && out == ""
}

// isGitHubURL reports whether the remote URL is a GitHub repository.
func isGitHubURL(remote string) bool {
	return strings.Contains(remote, "github.com")
}

func runGit(ctx context.Context, dir string, args ...string) error {
	tctx, cancel := context.WithTimeout(ctx, gitOpTimeout)
	defer cancel()
	cmd := exec.CommandContext(tctx, "git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

func runGitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	tctx, cancel := context.WithTimeout(ctx, gitOpTimeout)
	defer cancel()
	cmd := exec.CommandContext(tctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func runGitOutputInt(ctx context.Context, dir string, args ...string) (int, error) {
	out, err := runGitOutput(ctx, dir, args...)
	if err != nil {
		return 0, err
	}
	var n int
	if _, err := fmt.Sscanf(out, "%d", &n); err != nil {
		return 0, err
	}
	return n, nil
}

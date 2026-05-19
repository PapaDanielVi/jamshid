package gitvault

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
)

// CheckGhAuth verifies that `gh` CLI is installed and authenticated.
func CheckGhAuth() error {
	// Check if gh is installed
	if err := exec.Command("which", "gh").Run(); err != nil {
		return fmt.Errorf("`gh` CLI is not installed - install it from https://cli.github.com")
	}
	// Check if authenticated
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("`gh` is not authenticated - run `gh auth login` first")
	}
	return nil
}

// InitVault initializes a git vault at ~/.config/jamshid/ with the given remote.
func InitVault(remote string) error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	// git init
	if err := runGit(dir, "init"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	// add remote
	if err := runGit(dir, "remote", "add", "origin", remote); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}
	return nil
}

// SyncPush commits and pushes all changes to the vault remote.
func SyncPush() error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}

	// git add -A
	if err := runGit(dir, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	// git commit (only if there are changes)
	_ = runGit(dir, "commit", "-m", "sync: auto-update")
	// git push
	if err := runGit(dir, "push", "origin", "main"); err != nil {
		// Try master branch
		if err2 := runGit(dir, "push", "origin", "master"); err2 != nil {
			return fmt.Errorf("git push: %w", err)
		}
	}
	return nil
}

// SyncPull pulls latest changes from the vault remote.
func SyncPull() error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}

	if err := runGit(dir, "pull", "origin", "main"); err != nil {
		if err2 := runGit(dir, "pull", "origin", "master"); err2 != nil {
			return fmt.Errorf("git pull: %w", err)
		}
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

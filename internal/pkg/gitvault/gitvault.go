package gitvault

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/constants"
)

func CheckGhAuth() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("`gh` CLI is not installed - install it from https://cli.github.com")
	}
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("`gh` is not authenticated - run `gh auth login` first")
	}
	return nil
}

func InitVault(remote string) error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, constants.DefaultDirPerm); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := runGit(dir, "init"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	if err := runGit(dir, "remote", "add", "origin", remote); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}
	return nil
}

func SyncPush() error {
	dir, err := config.JamshidDir()
	if err != nil {
		return err
	}

	if err := runGit(dir, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Only commit if there are staged changes
	if err := runGit(dir, "diff", "--cached", "--quiet"); err != nil {
		// diff --cached --quiet exits 1 when there are differences
		_ = runGit(dir, "commit", "-m", constants.DefaultCommitMessage)
	}

	if err := runGit(dir, "push", "origin", "main"); err != nil {
		if err2 := runGit(dir, "push", "origin", "master"); err2 != nil {
			return fmt.Errorf("git push: %w", err)
		}
	}
	return nil
}

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

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitVault(t *testing.T) {
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", dir)
	defer func() { _ = os.Setenv("HOME", origHome) }()

	if err := InitVault("https://example.com/repo.git"); err != nil {
		t.Fatalf("InitVault: %v", err)
	}

	// Check .git exists
	gitDir := filepath.Join(dir, ".config/jamshid/.git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory not created")
	}
}

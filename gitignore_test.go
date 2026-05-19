package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitignore(t *testing.T) {
	dir := t.TempDir()

	// Test creating new .gitignore
	if err := EnsureGitignore(dir); err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(data), gitignoreEntry) {
		t.Error(".gitignore missing entry")
	}

	// Test idempotency - should not duplicate
	if err := EnsureGitignore(dir); err != nil {
		t.Fatalf("EnsureGitignore second: %v", err)
	}
	data, _ = os.ReadFile(filepath.Join(dir, ".gitignore"))
	count := strings.Count(string(data), gitignoreEntry)
	if count != 1 {
		t.Errorf("entry count = %d, want 1", count)
	}
}

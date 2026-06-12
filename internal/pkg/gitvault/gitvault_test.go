package gitvault_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PapaDanielVi/jamshid/internal/pkg/gitvault"
)

func TestInitVault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	ctx := context.Background()
	if err := gitvault.InitVault(ctx, "https://example.com/repo.git"); err != nil {
		t.Fatalf("InitVault: %v", err)
	}

	gitDir := filepath.Join(dir, ".config/jamshid/.git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory not created")
	}
}

func TestEnsureVaultGitignore(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	ctx := context.Background()
	if err := gitvault.EnsureVaultGitignore(ctx); err != nil {
		t.Fatalf("EnsureVaultGitignore: %v", err)
	}

	gitignorePath := filepath.Join(dir, ".config/jamshid/.gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	content := string(data)
	for _, entry := range []string{".credentials.json", "history/", "todos/"} {
		if !strings.Contains(content, entry) {
			t.Errorf(".gitignore missing entry %q", entry)
		}
	}

	// Call again to verify idempotency.
	if err := gitvault.EnsureVaultGitignore(ctx); err != nil {
		t.Fatalf("EnsureVaultGitignore second call: %v", err)
	}
	data2, _ := os.ReadFile(gitignorePath)
	if strings.Count(string(data2), ".credentials.json") != 1 {
		t.Error(".credentials.json duplicated in .gitignore")
	}
}

func gitConfig(t *testing.T, dir, key, val string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", dir, "config", "--local", key, val).CombinedOutput(); err != nil {
		t.Fatalf("git config %s=%s: %v: %s", key, val, err, out)
	}
}

func TestSyncPushPull(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	// Create a bare repo to act as the remote.
	bareDir := t.TempDir()
	if out, err := exec.Command("git", "init", "--bare", bareDir).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v: %s", err, out)
	}

	// Set up the pusher HOME.
	pusherHome := t.TempDir()
	t.Setenv("HOME", pusherHome)

	ctx := context.Background()
	if err := gitvault.InitVault(ctx, bareDir); err != nil {
		t.Fatalf("InitVault: %v", err)
	}

	vaultDir := filepath.Join(pusherHome, ".config/jamshid")
	gitConfig(t, vaultDir, "user.email", "test@test.com")
	gitConfig(t, vaultDir, "user.name", "Test")

	// Write a test file in the vault dir.
	testFile := filepath.Join(vaultDir, "test-profile.json")
	if err := os.WriteFile(testFile, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	if err := gitvault.SyncPush(ctx); err != nil {
		t.Fatalf("SyncPush: %v", err)
	}

	// Determine what branch was pushed by inspecting the bare repo's HEAD.
	branchOut, _ := exec.Command("git", "-C", bareDir, "for-each-ref", "--format=%(refname:short)", "refs/heads/").Output()
	if len(strings.TrimSpace(string(branchOut))) == 0 {
		t.Fatal("remote has no branches after SyncPush")
	}
	branch := strings.Fields(strings.TrimSpace(string(branchOut)))[0]

	// Verify the remote has the commit.
	out, err := exec.Command("git", "-C", bareDir, "log", branch, "--oneline").Output()
	if err != nil {
		t.Fatalf("git log on remote branch %q: %v", branch, err)
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		t.Errorf("remote has no commits after SyncPush")
	}

	// Set up a second HOME to pull into.
	pullerHome := t.TempDir()
	t.Setenv("HOME", pullerHome)

	if err := gitvault.InitVault(ctx, bareDir); err != nil {
		t.Fatalf("InitVault for puller: %v", err)
	}
	pullerVaultDir := filepath.Join(pullerHome, ".config/jamshid")
	gitConfig(t, pullerVaultDir, "user.email", "test@test.com")
	gitConfig(t, pullerVaultDir, "user.name", "Test")
	// Switch to the branch that exists on the remote.
	_ = exec.Command("git", "-C", pullerVaultDir, "checkout", "-b", branch).Run()

	if err := gitvault.SyncPull(ctx); err != nil {
		t.Fatalf("SyncPull: %v", err)
	}

	pulled := filepath.Join(pullerVaultDir, "test-profile.json")
	if _, err := os.Stat(pulled); os.IsNotExist(err) {
		t.Error("test-profile.json not present after SyncPull")
	}
}

func TestGetVaultStatus(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	dir := t.TempDir()
	t.Setenv("HOME", dir)

	ctx := context.Background()
	if err := gitvault.InitVault(ctx, "https://example.com/repo.git"); err != nil {
		t.Fatalf("InitVault: %v", err)
	}

	vaultDir := filepath.Join(dir, ".config/jamshid")
	gitConfig(t, vaultDir, "user.email", "test@test.com")
	gitConfig(t, vaultDir, "user.name", "Test")

	// Create an initial commit so symbolic-ref resolves.
	placeholder := filepath.Join(vaultDir, ".gitkeep")
	if err := os.WriteFile(placeholder, nil, 0644); err != nil {
		t.Fatalf("write placeholder: %v", err)
	}
	if out, err := exec.Command("git", "-C", vaultDir, "add", ".gitkeep").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v: %s", err, out)
	}
	if out, err := exec.Command("git", "-C", vaultDir, "commit", "-m", "init").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v: %s", err, out)
	}

	status, err := gitvault.GetVaultStatus(ctx)
	if err != nil {
		t.Fatalf("GetVaultStatus: %v", err)
	}
	if status.Remote == "" {
		t.Error("Remote is empty")
	}
	if status.Branch == "" {
		t.Error("Branch is empty")
	}
}

func TestCheckPrivateRepo_NoValidRepo(t *testing.T) {
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh not installed")
	}
	ctx := context.Background()
	// A clearly non-existent repo must always return an error.
	err := gitvault.CheckPrivateRepo(ctx, "https://github.com/nonexistent-owner-jamshid-test/nonexistent-repo-jamshid-test")
	if err == nil {
		t.Error("expected error for non-existent repo")
	}
}

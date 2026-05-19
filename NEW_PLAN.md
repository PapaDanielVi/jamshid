# Jamshid Implementation Plan

## Context
Build "jamshid" from scratch — an open-source Claude Code profile manager/switcher CLI tool. The repo is empty (docs only). The goal is to let developers switch between multiple Claude Code configurations (personal Anthropic, enterprise keys, OpenRouter tiers) across project directories without manual file copying or key leaks. Language: **Go 1.26.3** with **Bubble Tea** TUI framework.

---

## 1. Project Structure (Go Best Practices)

```
jamshid/
├── main.go                 # Entry point, CLI arg routing
├── config.go               # Config struct, load/save functions, profile link tracking
├── config_test.go          # Tests for config operations
├── profile.go              # Profile struct, CRUD functions, symlink logic
├── profile_test.go         # Tests for profile operations
├── gitvault.go             # Git sync (os/exec, no go-git)
├── gitvault_test.go        # Tests for git vault operations
├── gitignore.go            # .gitignore check and update logic
├── gitignore_test.go       # Tests for gitignore operations
├── hash.go                 # Directory hashing utility
├── hash_test.go            # Tests for hashing
├── models.go               # Anthropic + OpenRouter model definitions
├── models_test.go          # Tests for model lookup/search
├── tui/
│   ├── model.go            # Bubble Tea model (state, init, update, view)
│   ├── keys.go             # Key mappings and help text
│   ├── styles.go           # Lip Gloss styles
│   └── components.go       # Reusable TUI components (model selector, etc.)
├── .github/
│   └── workflows/
│       └── ci.yml          # GitHub Actions: test, build, lint
├── .golangci.yml           # Linter configuration
├── Makefile                # Build, test, lint targets
├── go.mod
├── go.sum
├── .gitignore
├── CLAUDE.md
├── LICENSE
├── README.md               # Professional README with badges
├── PLAN.md
└── NEW_PLAN.md
```

### Go Best Practices Followed
- Package-level documentation comments on all public types and functions
- Error handling: wrap errors with `fmt.Errorf("context: %w", err)` for traceability
- Tests co-located with source (`*_test.go`)
- `go vet` and `golangci-lint` in CI
- Table-driven tests where applicable
- `defer` for cleanup (file handles, temp dirs in tests)

### Runtime directory layout (`~/.config/jamshid/`)

```
~/.config/jamshid/
├── config.json              # Global config (profiles, global profile, linked dirs)
├── profiles/
│   ├── personal/
│   │   └── .claude/
│   │       ├── settings.json
│   │       └── settings.local.json
│   └── work/
│       └── .claude/
│           ├── settings.json
│           └── settings.local.json
└── (optional: git init here for vault sync)
```

---

## 2. Core Data Structures (`config.go`, `profile.go`, `models.go`)

```go
// hash.go
// DirHash returns a short hex hash (first 8 chars of SHA-256) of the given directory path.
func DirHash(path string) string

// models.go
type Model struct {
    ID          string `json:"id"`
    Name        string `json:"name"`        // Display name
    Provider    string `json:"provider"`    // "anthropic" or "openrouter"
    Description string `json:"description"`
}

var AnthropicModels = []Model{...}   // claude-opus-4-7, claude-sonnet-4-6, etc.
var OpenRouterModels = []Model{...}  // openrouter variants
func SearchModels(query string) []Model

// profile.go
type McpServer struct {
    Name    string   `json:"name"`
    Command string   `json:"command"`
    Args    []string `json:"args,omitempty"`
}

type Profile struct {
    Name         string            `json:"name"`
    EnvVars      map[string]string `json:"env_vars,omitempty"`
    ClaudeConfig map[string]any   `json:"claude_config,omitempty"`
    McpServers   []McpServer      `json:"mcp_servers,omitempty"`
    Model        string            `json:"model,omitempty"`       // Selected model ID
    Timeout      string            `json:"timeout,omitempty"`     // Request timeout
}

// config.go
type DirEntry struct {
    Path    string `json:"path"`    // Absolute directory path
    Hash    string `json:"hash"`    // Hash of path for unique ID
    Profile string `json:"profile"` // Profile name linked here
}

type Config struct {
    Version       string              `json:"version"`
    GlobalProfile string              `json:"global_profile,omitempty"`
    Profiles      map[string]Profile `json:"profiles,omitempty"`
    VaultRemote   string              `json:"vault_remote,omitempty"`
    LinkedDirs    map[string]DirEntry `json:"linked_dirs,omitempty"` // key = hash(dir)
}
```

Config stored at `~/.config/jamshid/config.json`. Profile files stored at `~/.config/jamshid/profiles/<name>/.claude/`.

---

## 3. Key Mechanisms

### Symlink Strategy (not copying)
When `jamshid local <profile>` is run:
1. Check cwd is a git repo (`git rev-parse --is-inside-work-tree`)
2. Compute hash of cwd: `hash := DirHash(absPath)`
3. Create symlink: `<cwd>/.claude -> ~/.config/jamshid/profiles/<name>/.claude`
   - If `.claude` already exists as a real directory, back it up to `.claude.bak`
   - If `.claude` already exists as a symlink, remove and replace
4. Register in config: `config.LinkedDirs[hash] = DirEntry{Path: absPath, Hash: hash, Profile: name}`
5. Update `.gitignore` in cwd to include `.claude/settings.local.json`

When checking active profile for cwd:
1. Compute hash of cwd
2. Look up `config.LinkedDirs[hash]` → if found, that's the active profile
3. If not found, fall back to `config.GlobalProfile`

### Profile Configuration (new feature)
When `jamshid` is run inside a repo with an already-linked profile:
1. Detect active profile via cwd hash lookup
2. Launch TUI in **configure mode** for that profile
3. Interactive model selector: searchable list of Anthropic + OpenRouter models (bubbles list + textinput for filtering)
4. Editable fields: model, timeout, env vars, MCP servers
5. Save updates → rewrite profile settings.json + `SyncPush()` if vault enabled

### Model Search in TUI
- Text input at top of model list for filtering
- Real-time filtering as user types (case-insensitive substring match on model ID and name)
- Navigate filtered list with ↑↓
- Enter selects model

### Directory Hash
- Use `sha256(absPath)` truncated to 8 hex chars as the directory key
- Works offline, no git remote needed

### .gitignore Management
When linking a profile to a project:
1. Check if cwd is a git repo
2. If `.gitignore` exists, check if it contains `.claude/settings.local.json`
3. If not present, append `.claude/settings.local.json` to `.gitignore`
4. If `.gitignore` doesn't exist, create it with that entry

---

## 4. Implementation Order

### Step 0: Project scaffolding
```bash
go mod init github.com/mk/jamshid
go get github.com/charmbracelet/bubbletea/v2 github.com/charmbracelet/lipgloss/v2 github.com/charmbracelet/bubbles/v2
```
Create `Makefile`:
```makefile
.PHONY: build test lint fmt vet all

build:
	go build -o jamshid .

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

all: fmt vet lint test build
```
Create `.golangci.yml` with standard Go linters.
Verify: `make build` produces binary.

### Step 1: Hash utility (`hash.go` + `hash_test.go`)
- `DirHash(path string) string` — SHA-256 of absolute path, truncated to 8 hex chars
- Test: known path produces deterministic hash; different paths produce different hashes
- Verify: `go test ./... -run TestDirHash`

### Step 2: Config loading/saving (`config.go` + `config_test.go`)
- `LoadConfig() (*Config, error)` — reads `~/.config/jamshid/config.json`, creates dir if missing
- `SaveConfig(cfg *Config) error` — writes JSON to disk
- Test: round-trip load/save, missing file creates defaults, invalid JSON returns error
- Verify: `go test ./... -run TestConfig`

### Step 3: Profile CRUD (`profile.go` + `profile_test.go`)
Functions: `AddProfile`, `DeleteProfile`, `GetProfile`, `ListProfiles`, `ProfileDir(name) string`
Profile storage: `~/.config/jamshid/profiles/<name>/.claude/settings.json`
Tests: add/delete/get/list profiles, profile dir creation
CLI (manual arg parsing, no Cobra):
- `jamshid add <name>` — interactive prompts, creates profile dir + settings.json
- `jamshid delete <name>` — removes profile dir
- `jamshid list` — print profiles
- Verify: `go test ./...` passes, `go run . list` shows empty

### Step 4: Models list (`models.go` + `models_test.go`)
- Define `Model` struct with ID, Name, Provider, Description
- Populate `AnthropicModels` with current models (claude-opus-4-7, claude-sonnet-4-6, claude-haiku-4-5, etc.)
- Populate `OpenRouterModels` with common OpenRouter model IDs
- `SearchModels(query string) []Model` — case-insensitive substring match on ID + Name
- Test: search returns correct results, empty query returns all
- Verify: `go test ./... -run TestModels`

### Step 5: Symlink binding (`profile.go` + `gitignore.go` + `gitignore_test.go` + `main.go`)
- `LinkProfile(cwd, profileName string) error` — symlink + register in config + update .gitignore
- `UnlinkProfile(cwd string) error` — remove symlink, clean up config
- `IsGitRepo(dir string) bool` — check `git rev-parse --is-inside-work-tree`
- `EnsureGitignore(cwd string) error` — add `.claude/settings.local.json` to .gitignore
- `GetActiveProfile(cwd string) (string, error)` — hash cwd, check LinkedDirs, fallback to GlobalProfile
- `jamshid local <profile>` — link profile to cwd (checks git repo first)
- `jamshid global <profile>` — set global fallback
- `jamshid unlink` — remove symlink from cwd
- Tests: symlink creation/removal, gitignore updates, active profile lookup
- Verify: `go test ./...`, run `jamshid local myprofile` in a git repo

### Step 6: Git Vault (`gitvault.go` + `gitvault_test.go`)
Uses `os/exec` to run git commands (no go-git lib):
- `InitVault(remote string)` — git init in `~/.config/jamshid/`, add remote
- `SyncPush()` — git add, commit, push after any change
- `SyncPull()` — git pull on startup if vault configured
- `jamshid vault init <url>` and `jamshid vault sync`
- Tests: mock git commands with `exec.Command` replacement or use `t.TempDir()` + real git
- Verify: `go test ./...`, configure test git repo, test vault init + sync

### Step 7: TUI (`tui/`)
Bubble Tea model with multiple views/states:
- **List view**: profiles list (bubbles list), status bar (active profile from cwd hash lookup, vault status), key hints
- **Configure view** (launched when cwd has linked profile): edit model (searchable selector), timeout, env vars, MCP servers
- **Model selector**: searchable list of Anthropic + OpenRouter models, text input for filtering, ↑↓ navigate, enter select
- **Form view**: text inputs for add/edit profile (bubbles textinput), masked input for API keys
- Keys:
  - List: ↑↓ navigate, enter = set global, `l` = link to cwd, `u` = unlink, `c` = configure (if active), `a` = add, `d` = delete, `q` = quit
  - Configure: tab cycles fields, text inputs for editable values, model field opens model selector
  - Model selector: type to filter, ↑↓ navigate, enter select, esc cancel
- Launch TUI when `jamshid` is run with no args; if cwd has linked profile, start in configure view for that profile
- Verify: `go run .` shows TUI, can navigate, configure profile, search/select models

### Step 8: GitHub Actions (`.github/workflows/ci.yml`)
```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26' }
      - run: make build
      - run: make test
      - uses: golangci/golangci-lint-action@v6
```

### Step 9: Professional README (`README.md`)
Include:
- **Badges**: Go Version, Build Status, Go Report Card, License, Release, Go Reference
- **Description**: What jamshid does, why it exists
- **Install**:
  ```bash
  # Go install (easiest)
  go install github.com/mk/jamshid@latest

  # Homebrew (macOS/Linux)
  brew tap mk/jamshid
  brew install jamshid
  ```
- **Quick Start**: copy-paste commands
- **Features**: bullet list (profile management, symlink switching, model selector, git vault)
- **CLI Reference**: table of commands
- **Contributing**: brief section
- **License**: reference

---

## 5. CLI Command Reference

| Command | Description |
|---------|-------------|
| `jamshid` | Launch TUI (configure mode if cwd has linked profile) |
| `jamshid add <name>` | Create new profile (interactive) |
| `jamshid delete <name>` | Delete profile |
| `jamshid list` | List all profiles with active status |
| `jamshid local <profile>` | Symlink profile to cwd (must be git repo) |
| `jamshid unlink` | Remove profile symlink from cwd |
| `jamshid global <profile>` | Set global fallback profile |
| `jamshid vault init <url>` | Configure git vault remote |
| `jamshid vault sync` | Trigger git sync |

---

## 6. Key Design Decisions
- **Symlinks, not copies** — `~/.config/jamshid/profiles/<name>/.claude/` linked into `<project>/.claude/`
- **Directory hash tracking** — `sha256(cwd)[:8]` as key in `LinkedDirs` map
- **Git repo required for local linking** — `jamshid local` checks `git rev-parse`
- **Auto .gitignore updates** — `.claude/settings.local.json` added automatically
- **Interactive model selector** — searchable list of Anthropic + OpenRouter models in TUI
- **Configure mode** — `jamshid` in a linked repo opens profile config (model, timeout, etc.)
- **No interfaces** — functions operating on structs, per simplicity-first principle
- **No Cobra** — manual arg parsing keeps it minimal
- **No go-git** — `os/exec` git commands avoid extra dependency
- **No encryption in v1** — git vault sync without encryption; noted for later
- **Config version fixed at "1"** — no migration logic needed yet
- **Tests co-located** — `*_test.go` files alongside source
- **CI via GitHub Actions** — test, build, lint on every push/PR

---

## 7. Verification

After each step:
1. `make build` succeeds with no errors
2. `make test` passes all tests (including new ones)
3. `make lint` passes with no issues
4. Config files and profile dirs are correct
5. Symlinks created correctly, `.gitignore` updated
6. TUI renders and responds to all key bindings
7. Model search/select works in TUI
8. Configure mode loads correct profile when run in linked repo
9. Git vault init + sync works with a test repo

End-to-end test:
```bash
# Create test git repo
mkdir -p /tmp/testproj && cd /tmp/testproj && git init

# Create and link profiles
go run . add personal
go run . add work
go run . global personal
go run . local work          # in /tmp/testproj (a git repo)

# Verify
ls -la /tmp/testproj/.claude  # should be symlink
cat /tmp/testproj/.gitignore   # should have .claude/settings.local.json
go run . list                  # should show work as active (linked)

# TUI - configure mode (in linked repo)
cd /tmp/testproj && go run .   # should start in configure mode for "work" profile
                             # can search/select model, change timeout, etc.

# TUI - list mode (not in a linked repo)
cd /tmp && go run .            # should start in list view

# Full CI check
make all                       # fmt, vet, lint, test, build all pass

# Install check
go install .                   # binary available as jamshid
jamshid list                  # works from anywhere
```

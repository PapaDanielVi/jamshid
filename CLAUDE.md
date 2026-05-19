## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

## 5. Jamshid-Specific Notes

**Testing**: `os.UserHomeDir()` doesn't respect `HOME` env var. Use `os.Getenv("HOME")` in `jamshidDir()` for testability.

**Bubble Tea**: Package name is `tea`, not `bubbletea`. Import as `tea "charm.land/bubbletea/v2"`. Bubbles uses `charm.land/bubbles/v2/list`. Lipgloss uses `charm.land/lipgloss/v2`.

**TUI Resize**: Handle `tea.WindowSizeMsg` in `Update()` and call `m.list.SetSize(w, h)`. Bubble Tea sends this automatically on startup and on SIGWINCH.

**Git Vault**: Check `gh` CLI auth with `exec.LookPath("gh")` (not `exec.Command("which", "gh")`) for cross-platform compatibility. Handle both "main" and "master" branch names.

**Symlinks**: When linking profiles, handle cases where `.claude` exists as real directory (backup to `.bak`) vs symlink (remove and replace). The symlink points to `settings.local.json`, not the `.claude` directory itself. MCP config files (`.mcp.json`, `mcp.json`, `mcp_servers.json`) are also symlinked from the project root.

**Linting**: `errcheck` linter requires checking return values of `os.Setenv`, `os.MkdirAll`, `os.Remove`, etc. Use `_ =` prefix if intentionally ignoring.

**Go Module Paths**: Bubble Tea v2 uses `charm.land/` paths. Current: `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`.

**Constants**: Shared constants (`.claude`, `settings.local.json`, `settings.json`, `config.json`, `profiles`, file/dir permissions, config version, commit message) live in `internal/pkg/constants`. Always use these instead of hardcoded values.

**Sentinel Errors**: `profile` package defines `ErrProfileNotFound`, `ErrProfileExists`, `ErrEmptyName`, `ErrSettingsExists`. `config` package defines `ErrConfigCorrupt`. Always wrap with `%w` so callers can use `errors.Is()`.

**Config Writes**: `SaveConfig` uses atomic writes (temp file + `os.Rename`) to prevent corruption. Never write directly to the config path.

**Version**: Build-time version is set via `-ldflags "-X main.Version=$(VERSION)"`. The `Makefile` extracts it from git tags. GoReleaser passes `{{ .Version }}`. Default is `"dev"`.

**MCP Configs**: Known MCP config file names are `.mcp.json`, `mcp.json`, `mcp_servers.json`. When adding a profile from an existing `.claude` directory, Jamshid also looks for these files in the project root (parent of `.claude`) and copies them into the profile directory. When linking/unlinking, these files are symlinked/removed alongside `settings.local.json`.

**Env Mode**: The `env` command prints `export CLAUDE_CONFIG_DIR=<path>` for use with `eval $(jamshid env <profile>)`. Claude Code reads `CLAUDE_CONFIG_DIR` to find its config. No symlinks needed. Requires a profile name argument.

**TUI Flow**: All commands must work through both CLI args and the TUI. The TUI uses multiple view states: `ViewCommands` (main menu), `ViewProfiles` (profile selection list), `ViewTextInput` (text input for add/vault init), and `ViewVaultSubcommands` (vault sub-command list). After the TUI quits, `main.go` reads the selected command/arg/subcmd from `SelectedCommand()` and calls `executeCommand`. Commands that need no args (list, unlink, help) quit immediately. Commands that need a profile (delete, link, env) transition to profile selection. Commands that need text input (add, vault init) transition to text input. The `textinput.Focus()` in bubbles v2 returns a `tea.Cmd` — use a `focusTextInput` flag and return the focus command from `Update` before dispatching to sub-handlers.

**Bubble Tea v2 Key Changes**: `tea.KeyMsg` → `tea.KeyPressMsg`. `View() string` → `View() tea.View` (use `tea.NewView(str)`). `tea.WithAltScreen()` removed — set `view.AltScreen = true` on the returned View. `list.Model.Update()` and `textinput.Model.Update()` use value receivers returning `(Model, Cmd)`. `textinput.Focus()` returns `tea.Cmd` (not void). `list.SetItems()` returns `tea.Cmd`.

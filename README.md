# Jamshid

<img src="docs/jamshid_icon.png" alt="jamshid" width="300"/>

![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Build Status](https://github.com/PapaDanielVi/jamshid/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/PapaDanielVi/jamshid)](https://goreportcard.com/report/github.com/PapaDanielVi/jamshid)
[![GoDoc](https://pkg.go.dev/badge/github.com/PapaDanielVi/jamshid)](https://pkg.go.dev/github.com/PapaDanielVi/jamshid)


Jamshid is a CLI tool for managing multiple Claude Code profiles. Switch between personal Anthropic, enterprise keys, and OpenRouter configurations across project directories without manual file copying or key leaks.

## Naming Philosophy

Jamshid (also spelled Jamshēd) was a mythical Persian king who possessed a magical cup called the "Jam-e-Jam" (Cup of Jamshid). This cup was said to reveal all the realms of the world, allowing the king to see everything at once. Much like the magical cup that showed all realms, Jamshid the tool gives you visibility and control over all your Claude Code configurations across different projects and environments.

## Features

- **Profile Management**: Create, delete, and list profiles
- **Symlink Switching**: Link profiles to project directories via symlinks (not copies)
- **MCP Server Support**: Automatically copies and links MCP server config files (`.mcp.json`, `mcp.json`, `mcp_servers.json`)
- **Env Mode**: Print `export CLAUDE_CONFIG_DIR=<path>` for use with `eval $(jamshid env <profile>)` — no symlinks needed
- **Git Vault**: Sync profiles across machines via git
- **Interactive TUI**: Full terminal UI with profile selection, text input, and sub-command navigation — all commands work through both CLI args and TUI

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap PapaDanielVi/tap
brew install jamshid
```

### Go Install

```bash
go install github.com/PapaDanielVi/jamshid@latest
```

### Binary Download

Download the latest release for your platform from the [Releases](https://github.com/PapaDanielVi/jamshid/releases) page.

## Quick Start

```bash
# Create profiles
jamshid add personal
jamshid add work

# List profiles
jamshid list

# Link profile to current directory (creates symlinks)
cd /path/to/project
jamshid link work

# Or use env mode (no symlinks — eval sets CLAUDE_CONFIG_DIR in current shell)
eval $(jamshid env work)
claude

# Unlink
jamshid unlink

# Show help
jamshid help

# Check version
jamshid --version
```

## CLI Reference

| Command                    | Description                                               |
| -------------------------- | --------------------------------------------------------- |
| `jamshid`                  | Launch interactive TUI                                    |
| `jamshid add <name>`       | Create new profile (imports settings + MCP configs)       |
| `jamshid delete <name>`    | Delete profile                                            |
| `jamshid list`             | List all profiles with their paths                        |
| `jamshid link [profile]`   | Link profile to cwd via symlinks (interactive)            |
| `jamshid unlink`           | Remove profile symlinks from cwd                          |
| `jamshid env <profile>`    | Set `CLAUDE_CONFIG_DIR` to the profile's config directory |
| `jamshid vault init <url>` | Configure git vault remote                                |
| `jamshid vault sync`       | Trigger git sync                                          |
| `jamshid version`          | Print version                                             |
| `jamshid help`             | Show help message                                         |

## Examples

### Import existing settings when creating a profile

```bash
cd /path/to/project/with/.claude/settings.local.json
jamshid add myproject
# Output: Found existing .claude/settings.local.json. Create profile from this? (y/n): y
# Output: Profile "myproject" created
```

When importing, Jamshid copies the entire `.claude` directory contents **and** looks for MCP config files (`.mcp.json`, `mcp.json`, `mcp_servers.json`) in the project root, copying them into the profile directory.

### Link a profile (symlink mode)

```bash
cd /path/to/project
jamshid link work
# Creates symlinks:
#   .claude/settings.local.json -> ~/.config/jamshid/profiles/work/.claude/settings.local.json
#   .mcp.json -> ~/.config/jamshid/profiles/work/.mcp.json  (if exists)
```

### Use env mode (no symlinks)

```bash
# Set CLAUDE_CONFIG_DIR in the current shell via eval
eval $(jamshid env work)
# Output: (exports CLAUDE_CONFIG_DIR to the current shell)

# Now run Claude Code — it will use the profile's config directory
claude

# Or use it in one line:
eval $(jamshid env personal) && claude
```

### Link a profile interactively

```bash
cd /path/to/project
jamshid link
# Output:
# Available profiles:
#   1: personal
#   2: work
# Select profile (number or name): 2
# Output: Linked profile "work" to /path/to/project
```

## Interactive TUI

Run `jamshid` without arguments to launch the interactive terminal UI:

```bash
jamshid
```

The TUI provides a full interactive experience for all commands:

- **Command Selection**: Navigate and select from available commands
- **Profile Selection**: For commands that need a profile (delete, link, env), a profile list is shown
- **Text Input**: For commands that need text input (add, vault init), a text input prompt is shown
- **Sub-command Navigation**: Vault commands show a sub-command list (init, sync)

All commands work identically through both the TUI and direct CLI arguments.

**TUI Navigation**:
- `↑/↓` or `j/k`: Navigate lists
- `enter`: Select item
- `esc`: Go back (from profile/text input views)
- `q` or `ctrl+c`: Quit

## How It Works

Jamshid offers two ways to use profiles:

### Symlink Mode (`link` / `unlink`)

1. Profiles are stored in `~/.config/jamshid/profiles/<name>/`
2. `jamshid link <profile>` creates symlinks:
   - `<cwd>/.claude/settings.local.json` → `~/.config/jamshid/profiles/<name>/.claude/settings.local.json`
   - `<cwd>/.mcp.json` → `~/.config/jamshid/profiles/<name>/.mcp.json` (if MCP configs exist)
3. The active profile for a directory is tracked via a hash of the directory path
4. `.gitignore` is automatically updated to exclude `settings.local.json`

### Env Mode (`env`)

1. `jamshid env <profile>` prints `export CLAUDE_CONFIG_DIR=<path>` to stdout
2. Use `eval $(jamshid env <profile>)` to set the variable in your current shell
3. Claude Code reads `CLAUDE_CONFIG_DIR` to find its config — no symlinks needed
4. This is ideal for users who prefer environment-based config switching

## Project Structure

```
jamshid/
├── cmd/jamshid/              # Entry point
├── internal/
│   ├── pkg/
│   │   ├── config/           # Config load/save (atomic writes)
│   │   ├── constants/        # Shared constants
│   │   ├── gitignore/        # .gitignore management
│   │   ├── gitvault/         # Git vault sync
│   │   ├── hash/             # Directory hashing
│   │   ├── models/           # Profile and MCP server types
│   │   └── profile/          # Profile CRUD, symlinks, MCP, env
│   └── tui/                  # Bubble Tea TUI
├── Formula/                  # Homebrew formula
├── Makefile                  # Build, test, lint
├── .goreleaser.yaml          # Release config
└── .github/workflows/        # CI/CD
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT

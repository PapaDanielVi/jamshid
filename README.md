# Jamshid

![Jamshid Icon](docs/jamshid_icon.png)

![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Build Status](https://github.com/PapaDanielVi/jamshid/actions/workflows/ci.yml/badge.svg)

Jamshid is a CLI tool for managing multiple Claude Code profiles. Switch between personal Anthropic, enterprise keys, and OpenRouter configurations across project directories without manual file copying or key leaks.

## Naming Philosophy

Jamshid (also spelled Jamshēd) was a mythical Persian king who possessed a magical cup called the "Jam-e-Jam" (Cup of Jamshid). This cup was said to reveal all the realms of the world, allowing the king to see everything at once. Much like the magical cup that showed all realms, Jamshid the tool gives you visibility and control over all your Claude Code configurations across different projects and environments.

## Features

- **Profile Management**: Create, delete, and list profiles
- **Symlink Switching**: Link profiles to project directories via symlinks (not copies)
- **MCP Server Support**: Automatically copies and links MCP server config files (`.mcp.json`, `mcp.json`, `mcp_servers.json`)
- **Env Mode**: Set `CLAUDE_CONFIG_DIR` environment variable to point Claude Code at a profile directory — no symlinks needed
- **Git Vault**: Sync profiles across machines via git
- **TUI**: Beautiful terminal UI built with Bubble Tea

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

# Or use env mode (no symlinks — sets CLAUDE_CONFIG_DIR)
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

| Command                    | Description                                           |
| -------------------------- | ----------------------------------------------------- |
| `jamshid`                  | Launch interactive TUI                                |
| `jamshid add <name>`       | Create new profile (imports settings + MCP configs)   |
| `jamshid delete <name>`    | Delete profile                                        |
| `jamshid list`             | List all profiles with their paths                    |
| `jamshid link [profile]`   | Link profile to cwd via symlinks (interactive)        |
| `jamshid unlink`           | Remove profile symlinks from cwd                      |
| `jamshid env [profile]`    | Print `CLAUDE_CONFIG_DIR` export for a profile        |
| `jamshid vault init <url>` | Configure git vault remote                            |
| `jamshid vault sync`       | Trigger git sync                                      |
| `jamshid version`          | Print version                                         |
| `jamshid help`             | Show help message                                     |

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
# Set CLAUDE_CONFIG_DIR for the current shell session
eval $(jamshid env work)
# Output: export CLAUDE_CONFIG_DIR=~/.config/jamshid/profiles/work/.claude

# Now run Claude Code — it will use the profile's config directory
claude

# Print all profile env vars
eval $(jamshid env)
# Output:
#   export CLAUDE_CONFIG_DIR=~/.config/jamshid/profiles/personal/.claude
#   export CLAUDE_CONFIG_DIR=~/.config/jamshid/profiles/work/.claude
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

1. `jamshid env <profile>` prints `export CLAUDE_CONFIG_DIR=~/.config/jamshid/profiles/<name>/.claude`
2. Use `eval $(jamshid env <profile>)` in your shell to set the variable
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

# Jamshid

![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Build Status](https://github.com/PapaDanielVi/jamshid/actions/workflows/ci.yml/badge.svg)

Jamshid is a CLI tool for managing multiple Claude Code profiles. Switch between personal Anthropic, enterprise keys, and OpenRouter configurations across project directories without manual file copying or key leaks.

## Naming Philosophy

Jamshid (also spelled Jamshid) was a mythical Persian king who possessed a magical cup called the "Jam-e-Jam" (Cup of Jamshid). This cup was said to reveal all the realms of the world, allowing the king to see everything at once. Much like the magical cup that showed all realms, Jamshid the tool gives you visibility and control over all your Claude Code configurations across different projects and environments.

## Features

- **Profile Management**: Create, delete, and list profiles
- **Symlink Switching**: Link profiles to project directories via symlinks (not copies)
- **Model Selector**: Interactive searchable list of Anthropic and OpenRouter models
- **Git Vault**: Sync profiles across machines via git
- **TUI**: Beautiful terminal UI built with Bubble Tea

## Installation

### Go Install

```bash
go install github.com/PapaDanielVi/jamshid@latest
```

### Homebrew (macOS/Linux)

```bash
brew tap PapaDanielVi/jamshid
brew install jamshid
```

## Quick Start

```bash
# Create profiles
jamshid add personal
jamshid add work

# Set global profile
jamshid global personal

# Link profile to current directory (must be git repo)
cd /path/to/project
jamshid local work

# List profiles
jamshid list

# Launch TUI
jamshid
```

## CLI Reference

| Command                    | Description                                           |
| -------------------------- | ----------------------------------------------------- |
| `jamshid`                  | Launch TUI (configure mode if cwd has linked profile) |
| `jamshid add <name>`       | Create new profile (interactive)                      |
| `jamshid delete <name>`    | Delete profile                                        |
| `jamshid list`             | List all profiles with active status                  |
| `jamshid local <profile>`  | Symlink profile to cwd (must be git repo)             |
| `jamshid unlink`           | Remove profile symlink from cwd                       |
| `jamshid global <profile>` | Set global fallback profile                           |
| `jamshid vault init <url>` | Configure git vault remote                            |
| `jamshid vault sync`       | Trigger git sync                                      |

## How It Works

Jamshid uses symlinks to switch between Claude Code configurations:

1. Profiles are stored in `~/.config/jamshid/profiles/<name>/`
2. When you run `jamshid local <profile>`, it creates a symlink: `<cwd>/.claude -> ~/.config/jamshid/profiles/<name>/.claude`
3. The active profile for a directory is tracked via a hash of the directory path
4. `.gitignore` is automatically updated to exclude `settings.local.json`

## Contributing

Contributions are welcome! Please feel free to submit a PR.

## License

MIT

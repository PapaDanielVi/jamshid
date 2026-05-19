package gitignore

import (
	"os"
	"path/filepath"
	"strings"
)

const gitignoreEntry = ".claude/settings.local.json"

// EnsureGitignore adds .claude/settings.local.json to .gitignore if not present.
func EnsureGitignore(cwd string) error {
	gitignore := filepath.Join(cwd, ".gitignore")

	var lines []string
	if data, err := os.ReadFile(gitignore); err == nil {
		lines = strings.Split(string(data), "\n")
	}

	// Check if entry already exists
	for _, line := range lines {
		if strings.TrimSpace(line) == gitignoreEntry {
			return nil
		}
	}

	// Append entry
	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if len(lines) > 0 && lines[len(lines)-1] != "" {
		_, _ = f.WriteString("\n")
	}
	if _, err := f.WriteString(gitignoreEntry + "\n"); err != nil {
		return err
	}
	return nil
}

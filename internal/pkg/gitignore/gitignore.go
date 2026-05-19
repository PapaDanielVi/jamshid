package gitignore

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/PapaDanielVi/jamshid/internal/pkg/constants"
)

const gitignoreEntry = constants.DirClaude + "/" + constants.FileSettingsLocal

func EnsureGitignore(cwd string) error {
	gitignore := filepath.Join(cwd, ".gitignore")

	var lines []string
	data, err := os.ReadFile(gitignore)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		lines = strings.Split(string(data), "\n")
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == gitignoreEntry {
			return nil
		}
	}

	f, err := os.OpenFile(gitignore, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constants.DefaultFilePerm)
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

// Package constants holds shared constants used across the jamshid codebase.
package constants

import "os"

const (
	DirClaude            = ".claude"
	FileSettingsLocal    = "settings.local.json"
	FileSettingsJSON     = "settings.json"
	FileConfigJSON       = "config.json"
	DirProfiles          = "profiles"
	ConfigVersion        = "1"
	DefaultCommitMessage = "sync: auto-update"
)

const (
	DefaultDirPerm  os.FileMode = 0755
	DefaultFilePerm os.FileMode = 0644
)

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PapaDanielVi/jamshid/internal/pkg/constants"
	"github.com/PapaDanielVi/jamshid/internal/pkg/models"
)

var ErrConfigCorrupt = errors.New("config file contains invalid JSON")

type DirEntry struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	Profile string `json:"profile"`
}

type Config struct {
	Version     string                    `json:"version"`
	Profiles    map[string]models.Profile `json:"profiles,omitempty"`
	VaultRemote string                    `json:"vault_remote,omitempty"`
	LinkedDirs  map[string]DirEntry       `json:"linked_dirs,omitempty"`
}

func JamshidDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
	}
	return filepath.Join(home, ".config/jamshid"), nil
}

func configPath() (string, error) {
	dir, err := JamshidDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, constants.FileConfigJSON), nil
}

func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), constants.DefaultDirPerm); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Version: constants.ConfigVersion, Profiles: make(map[string]models.Profile)}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigCorrupt, err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]models.Profile)
	}
	if cfg.LinkedDirs == nil {
		cfg.LinkedDirs = make(map[string]DirEntry)
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), constants.DefaultDirPerm); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, constants.DefaultFilePerm); err != nil {
		return fmt.Errorf("write config temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

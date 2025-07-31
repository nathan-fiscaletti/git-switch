package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"gopkg.in/yaml.v2"
)

var (
	ErrRepositoryNotFound = errors.New("repository not found")
)

const (
	StorageDirectory string = ".gitswitch"
)

type RepositoryConfig struct {
	Path           string   `yaml:"path"`
	PinnedBranches []string `yaml:"pinned-branches"`
	LastBranch     string   `yaml:"last-branch"`
}

type Config struct {
	Repositories        []RepositoryConfig `yaml:"repositories"`
	PinnedBranchPrefix  string             `yaml:"pinned-branch-prefix"`
	WindowSize          int                `yaml:"window-size"`
	PruneRemoteBranches bool               `yaml:"prune-remote-branches"`
}

func (c *Config) GetRepositoryConfig(path string) (*RepositoryConfig, error) {
	for _, rc := range c.Repositories {
		if strings.EqualFold(rc.Path, path) {
			return &rc, nil
		}
	}

	rc := RepositoryConfig{
		Path: path,
	}

	c.Repositories = append(c.Repositories, rc)

	return &rc, write(c)
}

func GetConfig() (*Config, error) {
	storagePath := configdir.LocalConfig(StorageDirectory)
	err := configdir.MakePath(storagePath) // Ensure it exists.
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(storagePath, "config")

	cfg := Config{
		PinnedBranchPrefix:  "★",
		Repositories:        []RepositoryConfig{},
		WindowSize:          10,
		PruneRemoteBranches: false,
	}

	if _, err := os.Stat(configFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		cfgBytes, err := yaml.Marshal(cfg)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(configFile, cfgBytes, 0660)
		if err != nil {
			return nil, err
		}

		return &cfg, nil
	}

	cfgData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.PinnedBranchPrefix == "" {
		cfg.PinnedBranchPrefix = "★"
	}

	if cfg.WindowSize == 0 {
		cfg.WindowSize = 10
	}

	return &cfg, nil
}

func write(cfg *Config) error {
	storagePath := configdir.LocalConfig(StorageDirectory)
	err := configdir.MakePath(storagePath) // Ensure it exists.
	if err != nil {
		return err
	}

	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	configFile := filepath.Join(storagePath, "config")

	return os.WriteFile(configFile, cfgBytes, 0660)
}

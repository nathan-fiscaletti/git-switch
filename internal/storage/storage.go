package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
)

const (
	StorageDirectory string = ".gitswitch"
)

type RepositoryConfig struct {
	Path           string   `yaml:"path"`
	PinnedBranches []string `yaml:"pinned-branches"`
}

type Config struct {
	Repositories       []RepositoryConfig `yaml:"repositories"`
	PinnedBranchPrefix string             `yaml:"pinned-branch-prefix"`
	WindowSize         int                `yaml:"window-size"`
}

func GetConfig() (*Config, error) {
	storagePath := configdir.LocalConfig(StorageDirectory)
	err := configdir.MakePath(storagePath) // Ensure it exists.
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(storagePath, "config")

	cfg := Config{
		PinnedBranchPrefix: "★",
		Repositories:       []RepositoryConfig{},
		WindowSize:         10,
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

func Pin(branch string) (*Config, error) {
	branch = strings.TrimSpace(branch)

	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}

	repositoryPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if _, idx, found := lo.FindIndexOf(cfg.Repositories, func(r RepositoryConfig) bool {
		return r.Path == repositoryPath
	}); found {
		if !lo.Contains(cfg.Repositories[idx].PinnedBranches, branch) {
			cfg.Repositories[idx].PinnedBranches = append(cfg.Repositories[idx].PinnedBranches, branch)
		}
	} else {
		cfg.Repositories = append(cfg.Repositories, RepositoryConfig{
			Path:           repositoryPath,
			PinnedBranches: []string{branch},
		})
	}

	return cfg, write(cfg)
}

var (
	ErrBranchNotPinned = errors.New("branch not pinned")
)

func Unpin(branch string) (*Config, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}

	repositoryPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Find the repository int he config.
	_, i, found := lo.FindIndexOf(cfg.Repositories, func(r RepositoryConfig) bool {
		return r.Path == repositoryPath
	})
	if !found {
		return nil, ErrBranchNotPinned
	}

	// Find the pinned branch
	_, j, found := lo.FindIndexOf(cfg.Repositories[i].PinnedBranches, func(f string) bool {
		return f == branch
	})

	if !found {
		return nil, ErrBranchNotPinned
	}

	// Remove the pinned branch
	cfg.Repositories[i].PinnedBranches = append(
		cfg.Repositories[i].PinnedBranches[:j],
		cfg.Repositories[i].PinnedBranches[j+1:]...,
	)

	return cfg, write(cfg)
}

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
	Path          string   `yaml:"path"`
	FocusBranches []string `yaml:"focus-branches"`
}

type Config struct {
	Repositories []RepositoryConfig `yaml:"repositories"`
}

func GetConfig() (*Config, error) {
	storagePath := configdir.LocalConfig(StorageDirectory)
	err := configdir.MakePath(storagePath) // Ensure it exists.
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(storagePath, "config")

	cfg := Config{
		Repositories: []RepositoryConfig{},
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

func Focus(branch string) (*Config, error) {
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
		if !lo.Contains(cfg.Repositories[idx].FocusBranches, branch) {
			cfg.Repositories[idx].FocusBranches = append(cfg.Repositories[idx].FocusBranches, branch)
		}
	} else {
		cfg.Repositories = append(cfg.Repositories, RepositoryConfig{
			Path:          repositoryPath,
			FocusBranches: []string{branch},
		})
	}

	return cfg, write(cfg)
}

var (
	ErrFocusBranchNotFound = errors.New("focus branch not found")
)

func Unfocus(branch string) (*Config, error) {
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
		return nil, ErrFocusBranchNotFound
	}

	// Find the focus branch
	_, j, found := lo.FindIndexOf(cfg.Repositories[i].FocusBranches, func(f string) bool {
		return f == branch
	})

	if !found {
		return nil, ErrFocusBranchNotFound
	}

	// Remove the focus branch
	cfg.Repositories[i].FocusBranches = append(
		cfg.Repositories[i].FocusBranches[:j],
		cfg.Repositories[i].FocusBranches[j+1:]...,
	)

	return cfg, write(cfg)
}

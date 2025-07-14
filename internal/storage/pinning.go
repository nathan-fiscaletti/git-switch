package storage

import (
	"errors"
	"strings"

	"github.com/nathan-fiscaletti/git-switch/internal/git"
	"github.com/samber/lo"
)

var (
	ErrBranchNotPinned = errors.New("branch not pinned")
)

func Pin(branch string) (*Config, error) {
	branch = strings.TrimSpace(branch)

	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}

	repositoryPath, err := git.GetRepositoryPath()
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

func Unpin(branch string) (*Config, error) {
	cfg, err := GetConfig()
	if err != nil {
		return nil, err
	}

	repositoryPath, err := git.GetRepositoryPath()
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

func ClearPins() error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	repositoryPath, err := git.GetRepositoryPath()
	if err != nil {
		return err
	}

	if _, idx, found := lo.FindIndexOf(cfg.Repositories, func(r RepositoryConfig) bool {
		return r.Path == repositoryPath
	}); found {
		cfg.Repositories[idx].PinnedBranches = []string{}
	}

	return write(cfg)
}

package storage

import (
	"strings"

	"github.com/nathan-fiscaletti/git-switch/internal/git"
	"github.com/samber/lo"
)

func SetLastBranch(branch string) (*Config, error) {
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
		cfg.Repositories[idx].LastBranch = branch
	} else {
		cfg.Repositories = append(cfg.Repositories, RepositoryConfig{
			Path:           repositoryPath,
			PinnedBranches: []string{},
			LastBranch:     branch,
		})
	}

	return cfg, write(cfg)
}

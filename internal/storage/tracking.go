package storage

import (
	"os"
	"strings"

	"github.com/samber/lo"
)

func SetLastBranch(branch string) (*Config, error) {
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

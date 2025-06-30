package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nathan-fiscaletti/git-switch/internal/git"
	"github.com/nathan-fiscaletti/git-switch/internal/storage"
	"github.com/nathan-fiscaletti/git-switch/pkg"
	"github.com/samber/lo"
)

func main() {
	err := git.ValidateGitInstallation()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	inRepo, _ := git.IsGitRepository()
	if !inRepo {
		fmt.Printf("error: %v\n", "not a git repository")
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		args := os.Args[2:]

		switch cmd {
		case "-x":
			switch args[0] {
			case "focus": // Maintain 'focus' for backwards compatibility.
				fallthrough
			case "pin":
				currentBranch, err := git.GetCurrentBranch()
				if err != nil {
					fmt.Printf("error: %v\n", err)
					os.Exit(1)
				}

				_, err = storage.Pin(currentBranch)
				if err != nil {
					fmt.Printf("error: %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			case "unfocus": // Maintain 'unfocus' for backwards compatibility
				fallthrough
			case "unpin":
				currentBranch, err := git.GetCurrentBranch()
				if err != nil {
					fmt.Printf("error: %v\n", err)
					os.Exit(1)
				}

				_, err = storage.Unpin(currentBranch)
				if err != nil {
					fmt.Printf("error: %v\n", err)
					os.Exit(1)
				}
				os.Exit(0)
			default:
				println("unknown command: %v", args[0])
				os.Exit(1)
			}
		default:
			err := git.ExecuteCheckout(strings.Join(os.Args[1:], " "))
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	branches, err := git.AllBranches()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := storage.GetConfig()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	repositoryPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	pinnedBranches := []string{}

	if repo, repoFound := lo.Find(cfg.Repositories, func(r storage.RepositoryConfig) bool {
		return r.Path == repositoryPath
	}); repoFound {
		pinnedBranches = repo.PinnedBranches
	}

	branchSelector, err := pkg.NewBranchSelector(pkg.BranchSelectorArguments{
		Branches:           branches,
		WindowSize:         10,
		SearchLabel:        "search branch",
		PinnedBranches:     pinnedBranches,
		PinnedBranchPrefix: cfg.PinnedBranchPrefix,
	})
	if err != nil {
		panic(err)
	}

	b, err := branchSelector.PickBranch()
	if err != nil {
		panic(err)
	}

	if len(b) > 0 {
		err = git.Checkout(b)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}

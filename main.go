package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nathan-fiscaletti/git-switch/internal/app"
	"github.com/nathan-fiscaletti/git-switch/internal/git"
	"github.com/nathan-fiscaletti/git-switch/internal/storage"
	"github.com/nathan-fiscaletti/git-switch/pkg"
)

func main() {
	var (
		pipeOutput = false
		pop        = false
	)

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		args := os.Args[2:]

		switch cmd {
		case "-x":
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
			case "pipe":
				pipeOutput = true
			case "pop":
				pop = true
			default:
				fmt.Printf("unknown internal command: %v\n", args[0])
				os.Exit(1)
			}
		case "--help":
			fallthrough
		case "-h":
			cmd := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
			fmt.Printf("%v: A fast, interactive terminal UI for switching between git branches.\n", cmd)
			println()
			println("Usage:")
			println()
			fmt.Printf("  %v [-h|-x <cmd>] ...\n", cmd)
			println()
			println("  - Run with no arguments for interactive `git checkout`")
			println("  - Run with arguments for regular `git checkout`")
			println("  - Run with `-x` for internal commands")
			println()
			println("Internal Commands:")
			println()
			println("  pin:   Pins the current branch")
			println("  unpin: Unpins the current branch")
			println("  pipe:  Pipes the selected branch name to stdout instead of checking it out")
			println("  pop:   Checks out the last branch you were in.")
			os.Exit(0)
		case "--version":
			fallthrough
		case "-v":
			v, err := app.GetVersionString()
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
			println(v)
			os.Exit(0)
		default:
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

			currentBranch, err := git.GetCurrentBranch()
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}

			_, err = storage.SetLastBranch(currentBranch)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}

			err = git.ExecuteCheckout(strings.Join(os.Args[1:], " "))
			if err != nil {
				fmt.Printf("error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

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

	cfg, err := storage.GetConfig()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	repositoryPath, err := git.GetRepositoryPath()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	repository, err := cfg.GetRepositoryConfig(repositoryPath)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	if pop {
		if repository.LastBranch == "" {
			fmt.Println("error: no branch to pop to")
			os.Exit(1)
		}

		_, err = storage.SetLastBranch(currentBranch)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}

		err = git.Checkout(repository.LastBranch)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if cfg.PruneRemoteBranches {
		err := git.PruneRemoteBranches()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	}

	branches, err := git.ListBranches()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	pinnedBranches := repository.PinnedBranches

	branchSelector, err := pkg.NewBranchSelector(pkg.BranchSelectorArguments{
		CurrentBranch:      currentBranch,
		Branches:           branches,
		WindowSize:         cfg.WindowSize,
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

	if pipeOutput {
		fmt.Println(b)
		os.Exit(0)
	}

	if len(b) > 0 {
		_, err = storage.SetLastBranch(currentBranch)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}

		err = git.Checkout(b)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(0)
}

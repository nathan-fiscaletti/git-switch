package pkg

import (
	"sync"

	"github.com/nathan-fiscaletti/git-switch/internal"
)

type BranchSelectorArguments struct {
	// The current branch that is checked out.
	CurrentBranch string
	// The list of branches to pick from.
	Branches []string
	// The pinned branches to display these will always be displayed at the
	// top of the list when able.
	PinnedBranches []string
	// PinnedBranchPrefix is the prefix to use before pinned branches.
	PinnedBranchPrefix string
	// The maximum number of branches to show at any given time.
	WindowSize int
	// The label to show in front of the search input.
	SearchLabel string
}

// Creates a new BranchSelector with the specified config
func NewBranchSelector(cfg BranchSelectorArguments) (*BranchSelector, error) {
	return &BranchSelector{cfg}, nil
}

type BranchSelector struct {
	cfg BranchSelectorArguments
}

// Present the branch selector to the user and return the selected branch.
func (b *BranchSelector) PickBranch() (string, error) {
	renderer, err := internal.NewRenderer(
		internal.RendererConfig{
			CurrentBranch:      b.cfg.CurrentBranch,
			Branches:           b.cfg.Branches,
			PinnedBranches:     b.cfg.PinnedBranches,
			WindowSize:         b.cfg.WindowSize,
			SearchLabel:        b.cfg.SearchLabel,
			PinnedBranchPrefix: b.cfg.PinnedBranchPrefix,
		},
	)
	if err != nil {
		return "", err
	}

	var (
		wg        sync.WaitGroup
		result    string
		resultErr error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Draw once to get started
		renderer.Draw()
		defer renderer.Finish()

		// Run updates
		for !renderer.IsDone() && resultErr == nil {
			err := renderer.Run(internal.SelectionHandler{
				OnSelect: func(v string) {
					result = v
				},
			})
			if err != nil {
				resultErr = err
				break
			}
		}
	}()

	wg.Wait()
	return result, resultErr
}

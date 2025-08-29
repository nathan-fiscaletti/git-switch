package pkg

import (
	"slices"
	"sync"

	"github.com/nathan-fiscaletti/git-switch/internal"
)

type BranchSelectorArguments struct {
	// The current branch that is checked out.
	CurrentBranch string
	// The list of branches to pick from.
	Branches []string
	// The pinned branches to display these will always be displayed at the
	// top of the list when able. Using a pointer allows the caller to
	// manage the pinned branches state.
	PinnedBranches *[]string
	// PinnedBranchPrefix is the prefix to use before pinned branches.
	PinnedBranchPrefix string
	// The maximum number of branches to show at any given time.
	WindowSize int
	// The label to show in front of the search input.
	SearchLabel string
	// Callback functions for pin/unpin operations
	OnPinBranch   func(branch string) error
	OnUnpinBranch func(branch string) error
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
				OnPin: func(branch string) error {
					// Only pin if the branch is not already pinned
					if !slices.Contains(*b.cfg.PinnedBranches, branch) {
						// Update the pinned branches list in-place
						*b.cfg.PinnedBranches = append(*b.cfg.PinnedBranches, branch)

						// Call the callback to handle storage operations
						if b.cfg.OnPinBranch != nil {
							if err := b.cfg.OnPinBranch(branch); err != nil {
								// Rollback the change if storage fails
								*b.cfg.PinnedBranches = (*b.cfg.PinnedBranches)[:len(*b.cfg.PinnedBranches)-1]
								return err
							}
						}
					}
					return nil
				},
				OnUnpin: func(branch string) error {
					// Only unpin if the branch is pinned
					if slices.Contains(*b.cfg.PinnedBranches, branch) {
						// Find and remove the branch from the pinned list
						pinnedBranches := *b.cfg.PinnedBranches
						for i, pinnedBranch := range pinnedBranches {
							if pinnedBranch == branch {
								// Store the old state for rollback
								oldPinnedBranches := make([]string, len(pinnedBranches))
								copy(oldPinnedBranches, pinnedBranches)

								// Remove the branch safely
								newPinnedBranches := make([]string, 0, len(pinnedBranches)-1)
								newPinnedBranches = append(newPinnedBranches, pinnedBranches[:i]...)
								newPinnedBranches = append(newPinnedBranches, pinnedBranches[i+1:]...)
								*b.cfg.PinnedBranches = newPinnedBranches

								// Call the callback to handle storage operations
								if b.cfg.OnUnpinBranch != nil {
									if err := b.cfg.OnUnpinBranch(branch); err != nil {
										// Rollback the change if storage fails
										*b.cfg.PinnedBranches = oldPinnedBranches
										return err
									}
								}
								break
							}
						}
					}
					return nil
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

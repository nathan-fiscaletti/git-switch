package internal

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/samber/lo"
)

type state struct {
	Input       string
	Branches    []string
	Selected    int
	Quit        bool
	WindowStart int
}

type Renderer struct {
	screen             tcell.Screen
	NormalStyle        tcell.Style
	BoldStyle          tcell.Style
	SelectedStyle      tcell.Style
	SelectedBold       tcell.Style
	InputStyle         tcell.Style
	CurrentBranchStyle tcell.Style
	WindowSize         int
	filter             func(input string) []string

	cfg         RendererConfig
	state       *state
	searchLabel string
}

type RendererConfig struct {
	Branches           []string
	PinnedBranches     *[]string
	WindowSize         int
	SearchLabel        string
	PinnedBranchPrefix string
	CurrentBranch      string
}

func NewRenderer(cfg RendererConfig) (*Renderer, error) {
	// Only include pinned branches if they are real branches.
	pinnedBranches := lo.Filter(*cfg.PinnedBranches, func(s string, _ int) bool {
		return lo.Contains(cfg.Branches, s)
	})

	// Remove pinned branches from normal branches
	normalBranches := lo.Filter(cfg.Branches, func(s string, _ int) bool {
		return !lo.Contains(pinnedBranches, s)
	})

	allBranches := append(pinnedBranches, normalBranches...)

	state := &state{
		Input:       "",
		Branches:    allBranches,
		Selected:    0,
		WindowStart: 0,
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := screen.Init(); err != nil {
		return nil, err
	}

	searchLabel := "search"
	if len(cfg.SearchLabel) > 0 {
		searchLabel = cfg.SearchLabel
	}

	renderer := &Renderer{
		NormalStyle:        tcell.StyleDefault,
		BoldStyle:          tcell.StyleDefault.Bold(true),
		SelectedStyle:      tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
		SelectedBold:       tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true),
		InputStyle:         tcell.StyleDefault.Foreground(tcell.ColorGreen),
		CurrentBranchStyle: tcell.StyleDefault.Foreground(tcell.ColorBlueViolet),
		WindowSize:         cfg.WindowSize,

		screen:      screen,
		state:       state,
		searchLabel: searchLabel,
		cfg:         cfg,
	}

	// Create the filter function that references the renderer's config
	filter := func(input string) []string {
		// Recalculate pinned and normal branches fresh each time using renderer's current config
		currentPinnedBranches := lo.Filter(*renderer.cfg.PinnedBranches, func(s string, _ int) bool {
			return lo.Contains(renderer.cfg.Branches, s)
		})

		// Remove pinned branches from normal branches
		currentNormalBranches := lo.Filter(renderer.cfg.Branches, func(s string, _ int) bool {
			return !lo.Contains(currentPinnedBranches, s)
		})

		// Create a fresh slice each time to avoid sharing issues
		allBranches := make([]string, 0, len(currentPinnedBranches)+len(currentNormalBranches))
		allBranches = append(allBranches, currentPinnedBranches...)
		allBranches = append(allBranches, currentNormalBranches...)

		if input == "" {
			// Return a copy to avoid slice sharing
			result := make([]string, len(allBranches))
			copy(result, allBranches)
			return result
		}

		// Filter all branches
		filtered := lo.Filter(allBranches, func(s string, _ int) bool {
			return strings.Contains(strings.ToLower(s), strings.ToLower(input))
		})

		// Return deduplicated result
		return lo.Uniq(filtered)
	}

	renderer.filter = filter
	return renderer, nil
}

func (r *Renderer) Draw() {
	r.screen.Clear()

	row := 0

	// 1. Draw hotkey instructions at the top
	col := 0
	orangeStyle := r.NormalStyle.Foreground(tcell.ColorOrange)
	dimStyle := r.NormalStyle.Dim(true)

	// CTRL+D in orange
	ctrlD := "CTRL+D"
	for _, ch := range ctrlD {
		r.screen.SetContent(col, row, ch, nil, orangeStyle)
		col++
	}

	// Rest of first instruction in dimmed text
	firstRest := ": Pin Selected Branch, "
	for _, ch := range firstRest {
		r.screen.SetContent(col, row, ch, nil, dimStyle)
		col++
	}

	// CTRL+U in orange
	ctrlU := "CTRL+U"
	for _, ch := range ctrlU {
		r.screen.SetContent(col, row, ch, nil, orangeStyle)
		col++
	}

	// Rest of second instruction in dimmed text
	secondRest := ": Unpin Selected Branch"
	for _, ch := range secondRest {
		r.screen.SetContent(col, row, ch, nil, dimStyle)
		col++
	}
	row++

	// 2. Empty line after hotkey instructions
	row++

	// 3. Draw current branch
	if r.cfg.CurrentBranch != "" {
		currentBranchLabel := fmt.Sprintf("checked out: %v", r.cfg.CurrentBranch)
		for i, ch := range currentBranchLabel {
			r.screen.SetContent(i, row, ch, nil, r.CurrentBranchStyle)
		}
		row++
	}

	// 4. Empty line after current branch
	row++

	// 5. Draw input at the next line
	inputPrompt := fmt.Sprintf("%v: %v", r.searchLabel, r.state.Input)
	for i, ch := range inputPrompt {
		r.screen.SetContent(i, row, ch, nil, r.InputStyle)
	}
	row++

	// 6. Empty line after input
	row++

	// 7. Draw list starting at the next line
	end := min(r.state.WindowStart+r.WindowSize, len(r.state.Branches))
	for i := r.state.WindowStart; i < end; i++ {
		item := r.state.Branches[i]
		isPinned := lo.Contains(*r.cfg.PinnedBranches, item)
		lowerItem := strings.ToLower(item)
		lowerInput := strings.ToLower(r.state.Input)
		start := strings.Index(lowerItem, lowerInput)
		col := 0
		isSelected := i == r.state.Selected
		style := r.NormalStyle
		bold := r.BoldStyle
		if isSelected {
			style = r.SelectedStyle
			bold = r.SelectedBold
		}
		// Render the pinned prefix in normal style, never selected/bold
		if isPinned {
			pinnedPrefix := fmt.Sprintf("%v ", r.cfg.PinnedBranchPrefix)
			for _, ch := range pinnedPrefix {
				r.screen.SetContent(col, row+i-r.state.WindowStart, ch, nil, r.NormalStyle)
				col++
			}
		}
		// Render the branch name (with selection/match logic)
		if r.state.Input != "" && start != -1 {
			// Before match
			for _, ch := range item[:start] {
				r.screen.SetContent(col, row+i-r.state.WindowStart, ch, nil, style)
				col++
			}
			// Match in bold
			for _, ch := range item[start : start+len(r.state.Input)] {
				r.screen.SetContent(col, row+i-r.state.WindowStart, ch, nil, bold)
				col++
			}
			// After match
			for _, ch := range item[start+len(r.state.Input):] {
				r.screen.SetContent(col, row+i-r.state.WindowStart, ch, nil, style)
				col++
			}
		} else {
			for _, ch := range item {
				r.screen.SetContent(col, row+i-r.state.WindowStart, ch, nil, style)
				col++
			}
		}
	}

	r.screen.Show()
}

type SelectionHandler struct {
	OnSelect func(string)
	OnPin    func(string) error
	OnUnpin  func(string) error
}

func (r *Renderer) Run(handler SelectionHandler) error {
	ev := r.screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEsc, tcell.KeyCtrlC:
			r.state.Quit = true
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(r.state.Input) > 0 {
				r.state.Input = r.state.Input[:len(r.state.Input)-1]
				r.state.Selected = 0
				r.state.WindowStart = 0
			}
		case tcell.KeyUp:
			if r.state.Selected > 0 {
				r.state.Selected--
				if r.state.Selected < r.state.WindowStart {
					r.state.WindowStart--
				}
			}
		case tcell.KeyDown:
			if r.state.Selected < len(r.state.Branches)-1 {
				r.state.Selected++
				if r.state.Selected >= r.state.WindowStart+r.WindowSize {
					r.state.WindowStart++
				}
			}
		case tcell.KeyEnter:
			if len(r.state.Branches) > 0 {
				r.state.Quit = true
				if handler.OnSelect != nil {
					handler.OnSelect(r.state.Branches[r.state.Selected])
					return nil
				}
			}
		case tcell.KeyCtrlD:
			// Pin the selected branch
			if len(r.state.Branches) > 0 && handler.OnPin != nil {
				selectedBranch := r.state.Branches[r.state.Selected]
				if err := handler.OnPin(selectedBranch); err != nil {
					return err
				}
				// Update the pinned branches list in the renderer
				if !lo.Contains(*r.cfg.PinnedBranches, selectedBranch) {
					*r.cfg.PinnedBranches = append(*r.cfg.PinnedBranches, selectedBranch)
					// Refresh the branch list and follow the pinned branch
					r.refreshBranchListWithSelection(selectedBranch, true)
					r.Draw()
					return nil
				}
			}
		case tcell.KeyCtrlU:
			// Unpin the selected branch
			if len(r.state.Branches) > 0 && handler.OnUnpin != nil {
				selectedBranch := r.state.Branches[r.state.Selected]
				if err := handler.OnUnpin(selectedBranch); err != nil {
					return err
				}
				// Update the pinned branches list in the renderer
				if idx := lo.IndexOf(*r.cfg.PinnedBranches, selectedBranch); idx != -1 {
					// Create a new slice to avoid memory corruption
					pinnedBranches := *r.cfg.PinnedBranches
					newPinnedBranches := make([]string, 0, len(pinnedBranches)-1)
					newPinnedBranches = append(newPinnedBranches, pinnedBranches[:idx]...)
					newPinnedBranches = append(newPinnedBranches, pinnedBranches[idx+1:]...)
					*r.cfg.PinnedBranches = newPinnedBranches

					// Find the next branch to select (stay in position instead of following)
					var nextBranch string
					if r.state.Selected+1 < len(r.state.Branches) {
						nextBranch = r.state.Branches[r.state.Selected+1]
					} else if r.state.Selected > 0 {
						nextBranch = r.state.Branches[r.state.Selected-1]
					}
					// Refresh the branch list and select the next branch
					r.refreshBranchListWithSelection(nextBranch, false)
					r.Draw()
					return nil
				}
			}
		default:
			if ev.Rune() != 0 {
				r.state.Input += string(ev.Rune())
				r.state.Selected = 0
				r.state.WindowStart = 0
			}
		}
		r.state.Branches = r.filter(r.state.Input)
		if r.state.Selected >= len(r.state.Branches) {
			r.state.Selected = len(r.state.Branches) - 1
		}
		if r.state.Selected < 0 {
			r.state.Selected = 0
		}
		if r.state.WindowStart > r.state.Selected {
			r.state.WindowStart = r.state.Selected
		}
		if r.state.WindowStart+r.WindowSize > len(r.state.Branches) {
			r.state.WindowStart = max(0, len(r.state.Branches)-r.WindowSize)
		}

		r.Draw()
	}

	return nil
}

func (r *Renderer) IsDone() bool {
	return r.state.Quit
}

func (r *Renderer) Finish() {
	r.screen.Fini()
}

func (r *Renderer) refreshBranchListWithSelection(targetBranch string, followBranch bool) {
	// Update the filter function to use the current branch configuration
	filter := func(input string) []string {
		// Recalculate pinned and normal branches fresh each time
		currentPinnedBranches := lo.Filter(*r.cfg.PinnedBranches, func(s string, _ int) bool {
			return lo.Contains(r.cfg.Branches, s)
		})

		// Remove pinned branches from normal branches
		currentNormalBranches := lo.Filter(r.cfg.Branches, func(s string, _ int) bool {
			return !lo.Contains(currentPinnedBranches, s)
		})

		// Create a fresh slice each time to avoid sharing issues
		allBranches := make([]string, 0, len(currentPinnedBranches)+len(currentNormalBranches))
		allBranches = append(allBranches, currentPinnedBranches...)
		allBranches = append(allBranches, currentNormalBranches...)

		if input == "" {
			// Return a copy to avoid slice sharing
			result := make([]string, len(allBranches))
			copy(result, allBranches)
			return result
		}

		// Filter all branches
		filtered := lo.Filter(allBranches, func(s string, _ int) bool {
			return strings.Contains(strings.ToLower(s), strings.ToLower(input))
		})

		// Return deduplicated result
		return lo.Uniq(filtered)
	}

	// Update the filter function and apply it
	r.filter = filter
	r.state.Branches = filter(r.state.Input)

	// Handle selection based on the followBranch parameter
	if followBranch && targetBranch != "" {
		// Try to select the target branch (follow behavior)
		if newIndex := lo.IndexOf(r.state.Branches, targetBranch); newIndex != -1 {
			r.state.Selected = newIndex
		}
	} else if !followBranch && targetBranch != "" {
		// Try to select the target branch, but if it doesn't exist, stay at current position
		if newIndex := lo.IndexOf(r.state.Branches, targetBranch); newIndex != -1 {
			r.state.Selected = newIndex
		}
		// If target branch doesn't exist, keep current selection position if possible
	}

	// Ensure the selected index is still valid
	if r.state.Selected >= len(r.state.Branches) {
		r.state.Selected = len(r.state.Branches) - 1
	}
	if r.state.Selected < 0 {
		r.state.Selected = 0
	}

	// Update window position to keep selected item visible
	if r.state.Selected < r.state.WindowStart {
		r.state.WindowStart = r.state.Selected
	}
	if r.state.Selected >= r.state.WindowStart+r.WindowSize {
		r.state.WindowStart = r.state.Selected - r.WindowSize + 1
	}
	if r.state.WindowStart < 0 {
		r.state.WindowStart = 0
	}
}

// min is kept here for local use
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max is kept here for local use
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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
	screen        tcell.Screen
	NormalStyle   tcell.Style
	BoldStyle     tcell.Style
	SelectedStyle tcell.Style
	SelectedBold  tcell.Style
	InputStyle    tcell.Style
	WindowSize    int
	filter        func(input string) []string

	cfg         RendererConfig
	state       *state
	searchLabel string
}

type RendererConfig struct {
	Branches           []string
	PinnedBranches     []string
	WindowSize         int
	SearchLabel        string
	PinnedBranchPrefix string
}

func NewRenderer(cfg RendererConfig) (*Renderer, error) {
	// Only include pinned branches if they are real branches.
	pinnedBranches := lo.Filter(cfg.PinnedBranches, func(s string, _ int) bool {
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

	filter := func(input string) []string {
		if input == "" && len(pinnedBranches) > 0 {
			// Show pinned branches at top, then the rest (deduped)
			return allBranches
		}
		// Filter all branches (deduped)
		filtered := lo.Filter(allBranches, func(s string, _ int) bool {
			return strings.Contains(strings.ToLower(s), strings.ToLower(input))
		})
		// Remove duplicates in case pinned and normal overlap in search
		return lo.Uniq(filtered)
	}

	searchLabel := "search"
	if len(cfg.SearchLabel) > 0 {
		searchLabel = cfg.SearchLabel
	}

	return &Renderer{
		NormalStyle:   tcell.StyleDefault,
		BoldStyle:     tcell.StyleDefault.Bold(true),
		SelectedStyle: tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
		SelectedBold:  tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true),
		InputStyle:    tcell.StyleDefault.Foreground(tcell.ColorGreen),
		WindowSize:    cfg.WindowSize,

		screen:      screen,
		filter:      filter,
		state:       state,
		searchLabel: searchLabel,
		cfg:         cfg,
	}, nil
}

func (r *Renderer) Draw() {
	r.screen.Clear()

	// Draw input at the top
	inputPrompt := fmt.Sprintf("%v: %v", r.searchLabel, r.state.Input)
	for i, ch := range inputPrompt {
		r.screen.SetContent(i, 0, ch, nil, r.InputStyle)
	}
	// Draw list starting at row 1
	end := min(r.state.WindowStart+r.WindowSize, len(r.state.Branches))
	for i := r.state.WindowStart; i < end; i++ {
		item := r.state.Branches[i]
		isPinned := lo.Contains(r.cfg.PinnedBranches, item)
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
				r.screen.SetContent(col, i-r.state.WindowStart+1, ch, nil, r.NormalStyle)
				col++
			}
		}
		// Render the branch name (with selection/match logic)
		if r.state.Input != "" && start != -1 {
			// Before match
			for _, ch := range item[:start] {
				r.screen.SetContent(col, i-r.state.WindowStart+1, ch, nil, style)
				col++
			}
			// Match in bold
			for _, ch := range item[start : start+len(r.state.Input)] {
				r.screen.SetContent(col, i-r.state.WindowStart+1, ch, nil, bold)
				col++
			}
			// After match
			for _, ch := range item[start+len(r.state.Input):] {
				r.screen.SetContent(col, i-r.state.WindowStart+1, ch, nil, style)
				col++
			}
		} else {
			for _, ch := range item {
				r.screen.SetContent(col, i-r.state.WindowStart+1, ch, nil, style)
				col++
			}
		}
	}
	r.screen.Show()
}

type SelectionHandler struct {
	OnSelect func(string)
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

// min is kept here for local use
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

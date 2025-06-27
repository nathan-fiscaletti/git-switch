package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nathan-fiscaletti/git-switch/internal/git"
	"github.com/samber/lo"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	things, err := git.RemoteBranches(context.Background())
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer screen.Fini()

	input := ""
	quit := false
	selected := 0
	windowStart := 0
	windowSize := 10

	normalStyle := tcell.StyleDefault
	boldStyle := tcell.StyleDefault.Bold(true)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	selectedBoldStyle := selectedStyle.Bold(true)
	inputStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)

	draw := func(matches []string, input string, selected, windowStart, windowSize int) {
		screen.Clear()
		// Draw input at the top
		inputPrompt := "search: " + input
		for i, r := range inputPrompt {
			screen.SetContent(i, 0, r, nil, inputStyle)
		}
		// Draw list starting at row 1
		end := min(windowStart+windowSize, len(matches))
		for i := windowStart; i < end; i++ {
			item := matches[i]
			lowerItem := strings.ToLower(item)
			lowerInput := strings.ToLower(input)
			start := strings.Index(lowerItem, lowerInput)
			col := 0
			isSelected := i == selected
			style := normalStyle
			bold := boldStyle
			if isSelected {
				style = selectedStyle
				bold = selectedBoldStyle
			}
			if input != "" && start != -1 {
				// Before match
				for _, r := range item[:start] {
					screen.SetContent(col, i-windowStart+1, r, nil, style)
					col++
				}
				// Match in bold
				for _, r := range item[start : start+len(input)] {
					screen.SetContent(col, i-windowStart+1, r, nil, bold)
					col++
				}
				// After match
				for _, r := range item[start+len(input):] {
					screen.SetContent(col, i-windowStart+1, r, nil, style)
					col++
				}
			} else {
				for _, r := range item {
					screen.SetContent(col, i-windowStart+1, r, nil, style)
					col++
				}
			}
		}
		screen.Show()
	}

	filter := func(input string) []string {
		if input == "" {
			return things
		}
		return lo.Filter(things, func(s string, _ int) bool {
			return strings.Contains(strings.ToLower(s), strings.ToLower(input))
		})
	}

	matches := filter(input)
	draw(matches, input, selected, windowStart, windowSize)

	go func() {
		for !quit {
			ev := screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEsc, tcell.KeyCtrlC:
					quit = true
					screen.Fini()
					return
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					if len(input) > 0 {
						input = input[:len(input)-1]
						selected = 0
						windowStart = 0
					}
				case tcell.KeyUp:
					if selected > 0 {
						selected--
						if selected < windowStart {
							windowStart--
						}
					}
				case tcell.KeyDown:
					if selected < len(matches)-1 {
						selected++
						if selected >= windowStart+windowSize {
							windowStart++
						}
					}
				case tcell.KeyEnter:
					if len(matches) > 0 {
						screen.Fini()
						err := git.Checkout(context.Background(), matches[selected])
						if err != nil {
							fmt.Printf("error: %v\n", err)
							os.Exit(1)
						}
						os.Exit(0)
					}
				default:
					if ev.Rune() != 0 {
						input += string(ev.Rune())
						selected = 0
						windowStart = 0
					}
				}
				matches = filter(input)
				if selected >= len(matches) {
					selected = len(matches) - 1
				}
				if selected < 0 {
					selected = 0
				}
				if windowStart > selected {
					windowStart = selected
				}
				if windowStart+windowSize > len(matches) {
					windowStart = max(0, len(matches)-windowSize)
				}
				draw(matches, input, selected, windowStart, windowSize)
			}
		}
	}()

	for !quit {
		time.Sleep(100 * time.Millisecond)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Package menu implements the main feature-selection screen for OmniClean.
package menu

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Selection identifies which feature the user chose.
type Selection int

const (
	SelectNone      Selection = iota
	SelectUninstall           // package uninstall TUI
	SelectAnalyze             // disk analyzer TUI
	SelectPurge               // project artifact purge TUI
	SelectQuit
)

type menuItem struct {
	title string
	desc  string
}

var menuItems = []menuItem{
	{"Uninstall Packages", "Search and remove packages across all managers"},
	{"Analyze Disk", "Explore disk usage and trash large files"},
	{"Purge Project Artifacts", "Clean node_modules, target, .venv and more"},
	{"Quit", "Exit OmniClean"},
}

// App is the BubbleTea model for the main menu.
type App struct {
	cursor   int
	selected Selection
	width    int
	height   int
}

// New returns a fresh App with cursor at position 0.
func New() *App {
	return &App{selected: SelectNone}
}

// Run launches the menu program and returns the user's selection.
func Run() (Selection, error) {
	p := tea.NewProgram(New(), tea.WithContext(context.Background()))
	final, err := p.Run()
	if err != nil {
		return SelectQuit, err
	}
	if m, ok := final.(*App); ok {
		return m.selected, nil
	}
	return SelectQuit, nil
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down", "j":
			if a.cursor < len(menuItems)-1 {
				a.cursor++
			}
		case "enter", " ":
			a.selected = Selection(a.cursor + 1)
			return a, tea.Quit
		case "q", "ctrl+c", "esc":
			a.selected = SelectQuit
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *App) View() tea.View {
	content := a.render()
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (a *App) render() string {
	header := boxStyle.Render(fmt.Sprintf(
		"%s\n%s",
		titleStyle.Render("✦ OmniClean"),
		subtitleStyle.Render("Unified system cleanup toolkit"),
	))

	var items strings.Builder
	for i, item := range menuItems {
		var pointer, title, desc string
		if i == a.cursor {
			pointer = cursorStyle.Render("▸ ")
			title = activeTitleStyle.Render(item.title)
			desc = activeDescStyle.Render("  " + item.desc)
		} else {
			pointer = "  "
			title = inactiveTitleStyle.Render(item.title)
			desc = inactiveDescStyle.Render("  " + item.desc)
		}
		items.WriteString(fmt.Sprintf("%s%s\n%s\n\n", pointer, title, desc))
	}

	help := helpStyle.Render("↑/↓  navigate · enter  select · q  quit")
	content := fmt.Sprintf("%s\n\n%s%s", header, items.String(), help)

	if a.width <= 0 {
		return content
	}
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

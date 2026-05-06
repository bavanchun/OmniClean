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

// menuItems lists the primary actions. Quit is intentionally NOT here:
// it is a key-only affordance (q / esc) surfaced by the help footer.
var menuItems = []menuItem{
	{"Uninstall Packages", "Search and remove packages across all managers"},
	{"Analyze Disk", "Explore disk usage and trash large files"},
	{"Purge Project Artifacts", "Clean node_modules, target, .venv and more"},
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

// renderCard composes a single action card with an accent bar, index badge,
// title, and description. Active cards use a gradient border via Lipgloss v2's
// BorderForegroundBlend; idle cards use a solid muted border.
func (a *App) renderCard(i int, item menuItem) string {
	style := cardIdle
	bar := barIdleStyle.Render(" ")
	idxStyle := indexIdleStyle
	titleStyle := inactiveTitleStyle
	descStyle := inactiveDescStyle
	if i == a.cursor {
		style = cardActive
		bar = barActiveStyle.Render("▌")
		idxStyle = indexActiveStyle
		titleStyle = activeTitleStyle
		descStyle = activeDescStyle
	}
	titleRow := fmt.Sprintf("%s %s  %s",
		bar,
		idxStyle.Render(fmt.Sprintf("[%d]", i+1)),
		titleStyle.Render(item.title),
	)
	descRow := descStyle.Render("     " + item.desc)
	return style.Render(titleRow + "\n" + descRow)
}

// renderBrandPanel composes the left-hand brand panel: ASCII banner,
// product line, tagline, and version meta. Width is implicit in the
// banner glyphs; the rounded border tracks them.
func (a *App) renderBrandPanel() string {
	var b strings.Builder
	for _, line := range bannerLines {
		b.WriteString(bannerStyle.Render(line))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	b.WriteString(titleStyle.Render("✦ OmniClean"))
	b.WriteByte('\n')
	b.WriteString(brandTaglineStyle.Render("Unified cleanup toolkit"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(brandMetaStyle.Render("Use ↑/↓ to navigate"))
	return brandPanelStyle.Render(b.String())
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
	cards := make([]string, len(menuItems))
	for i, it := range menuItems {
		cards[i] = a.renderCard(i, it)
	}
	right := lipgloss.JoinVertical(lipgloss.Left, cards...)
	left := a.renderBrandPanel()
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right)

	help := helpStyle.Render("↑/↓ navigate · 1-3 jump · ↵ select · q quit")
	content := lipgloss.JoinVertical(lipgloss.Center, body, "", help)

	if a.width <= 0 {
		return content
	}
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

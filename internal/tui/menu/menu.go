// Package menu implements the main feature-selection screen for OmniClean.
package menu

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// Selection identifies which feature the user chose.
type Selection int

const (
	SelectNone          Selection = iota
	SelectUninstall                // package uninstall TUI
	SelectAnalyze                  // disk analyzer TUI
	SelectPurge                    // project artifact purge TUI
	SelectUninstallApps            // macOS app bundle uninstall TUI (darwin only)
	SelectQuit
)

type menuItem struct {
	title string
	desc  string
}

// buildMenuItems returns the primary action list. Quit is intentionally NOT
// here — it is a key-only affordance (q / esc) surfaced by the help footer.
// On Darwin an extra entry for app-bundle uninstall is appended at runtime.
func buildMenuItems() []menuItem {
	items := []menuItem{
		{"Uninstall Packages", "Search and remove packages across all managers"},
		{"Analyze Disk", "Explore disk usage and trash large files"},
		{"Purge Project Artifacts", "Clean node_modules, target, .venv and more"},
	}
	if runtime.GOOS == "darwin" {
		items = append(items, menuItem{"Uninstall Apps", "Remove .app bundles and orphan files"})
	}
	return items
}

// Options tunes the menu's behavior at construction time.
type Options struct {
	// Fancy enables animated effects (spinner banner star, rotating
	// gradient offset on the active card border). Off by default —
	// the menu is byte-identical to the static version when false.
	Fancy bool
}

// tickMsg drives the rotating border-blend offset under --fancy.
type tickMsg time.Time

// blendTickInterval keeps the gradient rotation calm enough to stay
// well below 2% CPU on Apple Silicon idle.
const blendTickInterval = 150 * time.Millisecond

func tickCmd() tea.Cmd {
	return tea.Tick(blendTickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// App is the BubbleTea model for the main menu.
type App struct {
	cursor   int
	selected Selection
	items    []menuItem
	width    int
	height   int
	keys     keyMap
	help     help.Model
	fancy    bool
	spin     spinner.Model
	tick     int
}

// New returns a fresh App with cursor at position 0.
func New(opts Options) *App {
	s := spinner.New(spinner.WithSpinner(spinner.Spinner{
		Frames: []string{"✦", "✧", "✦", "✧"},
		FPS:    200 * time.Millisecond,
	}))
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary))
	return &App{
		selected: SelectNone,
		items:    buildMenuItems(),
		keys:     defaultKeys(),
		help:     help.New(),
		fancy:    opts.Fancy,
		spin:     s,
	}
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
		if a.fancy {
			style = style.BorderForegroundBlendOffset(a.tick)
		}
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
	star := "✦"
	if a.fancy {
		star = a.spin.View()
	}
	b.WriteString(titleStyle.Render(star + " OmniClean"))
	b.WriteByte('\n')
	b.WriteString(brandTaglineStyle.Render("Unified cleanup toolkit"))
	b.WriteByte('\n')
	b.WriteByte('\n')
	b.WriteString(brandMetaStyle.Render("Use ↑/↓ to navigate"))
	return brandPanelStyle.Render(b.String())
}

// Run launches the menu program and returns the user's selection.
func Run(opts Options) (Selection, error) {
	p := tea.NewProgram(New(opts), tea.WithContext(context.Background()))
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
	if a.fancy {
		return tea.Batch(a.spin.Tick, tickCmd())
	}
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.SetWidth(msg.Width)

	case spinner.TickMsg:
		if !a.fancy {
			return a, nil
		}
		var cmd tea.Cmd
		a.spin, cmd = a.spin.Update(msg)
		return a, cmd

	case tickMsg:
		if !a.fancy {
			return a, nil
		}
		a.tick++
		return a, tickCmd()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Up):
			if a.cursor > 0 {
				a.cursor--
			}
		case key.Matches(msg, a.keys.Down):
			if a.cursor < len(a.items)-1 {
				a.cursor++
			}
		case key.Matches(msg, a.keys.Select):
			a.selected = Selection(a.cursor + 1)
			return a, tea.Quit
		case key.Matches(msg, a.keys.Jump):
			idx := int(msg.String()[0] - '1')
			if idx < len(a.items) {
				a.cursor = idx
				a.selected = Selection(idx + 1)
				return a, tea.Quit
			}
		case key.Matches(msg, a.keys.Help):
			a.help.ShowAll = !a.help.ShowAll
		case key.Matches(msg, a.keys.Quit):
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

// minTwoColWidth is the threshold below which the menu collapses
// to a single-column stack (banner above cards) so nothing wraps or
// overflows on narrow terminals.
const minTwoColWidth = 80

func (a *App) render() string {
	cards := make([]string, len(a.items))
	for i, it := range a.items {
		cards[i] = a.renderCard(i, it)
	}
	right := lipgloss.JoinVertical(lipgloss.Left, cards...)
	left := a.renderBrandPanel()

	var body string
	if a.width > 0 && a.width < minTwoColWidth {
		body = lipgloss.JoinVertical(lipgloss.Center, left, "", right)
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right)
	}

	footer := a.help.View(a.keys)
	content := lipgloss.JoinVertical(lipgloss.Center, body, "", footer)

	if a.width <= 0 {
		return content
	}
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

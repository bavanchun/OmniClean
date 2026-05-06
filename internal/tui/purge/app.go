// Package purge wraps the bubbletea program that drives the
// `omniclean purge` command. Phase 2.7 implements loading + list states;
// Phase 2.8 layers the confirm, delete, and result states on top.
package purge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	corepurge "github.com/bavanchun/OmniClean/internal/purge"
	"github.com/bavanchun/OmniClean/internal/tui/components"
	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// Config captures everything the cobra command hands off to the TUI.
type Config struct {
	Roots     []string
	Options   corepurge.Options
	DryRun    bool
	NoConfirm bool
}

type viewState int

const (
	stateLoading viewState = iota
	stateList
	stateConfirm
	stateDeleting
	stateResult
)

// App is the root Bubbletea model.
type App struct {
	cfg     Config
	state   viewState
	theme   theme.Styles
	spinner spinner.Model

	width, height int

	// scan results
	targets  []corepurge.Target
	selected map[string]bool // path -> selected
	cursor   int
	scanErr  error

	// post-delete
	deleted []corepurge.Target
	failed  map[string]error
}

// New constructs the App from the supplied Config.
func New(cfg Config) *App {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary))
	return &App{
		cfg:      cfg,
		state:    stateLoading,
		theme:    theme.New(),
		spinner:  sp,
		selected: map[string]bool{},
		failed:   map[string]error{},
	}
}

// Run starts the TUI program.
func (a *App) Run(ctx context.Context) error {
	p := tea.NewProgram(a, tea.WithContext(ctx))
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("run purge tui: %w", err)
	}
	return nil
}

// EditPaths is the future paths editor; still a placeholder.
func EditPaths(_ string, _ []string) error {
	return errors.New("purge --paths editor lands in Phase 2.9")
}

// --- Bubbletea contract ---

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.spinner.Tick, a.startScan())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyPressMsg:
		return a.handleKey(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	case scanDoneMsg:
		a.scanErr = msg.Err
		a.targets = msg.Targets
		// Pre-select non-recent entries.
		for _, t := range msg.Targets {
			if !t.Recent {
				a.selected[t.Path] = true
			}
		}
		a.state = stateList
		return a, nil
	}
	return a, nil
}

func (a *App) View() tea.View {
	var content string
	switch a.state {
	case stateLoading:
		content = a.viewLoading()
	case stateList:
		content = a.viewList()
	default:
		content = a.theme.Subtle.Render("(state pending — confirm/delete views land in 2.8)")
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// --- Commands ---

func (a *App) startScan() tea.Cmd {
	cfg := a.cfg
	return func() tea.Msg {
		targets, err := corepurge.NewWalker().Scan(context.Background(), cfg.Roots, cfg.Options)
		return scanDoneMsg{Targets: targets, Err: err}
	}
}

// --- Key handling ---

func (a *App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	}
	if a.state != stateList || len(a.targets) == 0 {
		return a, nil
	}
	switch msg.String() {
	case "up", "k":
		if a.cursor > 0 {
			a.cursor--
		}
	case "down", "j":
		if a.cursor < len(a.targets)-1 {
			a.cursor++
		}
	case " ", "x":
		t := a.targets[a.cursor]
		a.selected[t.Path] = !a.selected[t.Path]
	case "a":
		// Toggle select-all (skip Recent if currently fully selected).
		allOn := true
		for _, t := range a.targets {
			if !a.selected[t.Path] {
				allOn = false
				break
			}
		}
		for _, t := range a.targets {
			a.selected[t.Path] = !allOn && !t.Recent
		}
	}
	return a, nil
}

// --- Views ---

func (a *App) viewLoading() string {
	header := a.theme.Subtitle.Render("✦  OmniClean purge  ✦")
	body := fmt.Sprintf("  %s  %s",
		a.spinner.View(),
		a.theme.Body.Render("Scanning project artifacts in "+strings.Join(a.cfg.Roots, ", ")),
	)
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "ctrl+c", Action: "cancel"},
	})
	panel := components.Panel(a.theme, " Purge ",
		lipgloss.JoinVertical(lipgloss.Left, header, "", body),
		components.PanelOpts{Width: a.width - 4, Accent: true})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func (a *App) viewList() string {
	if a.scanErr != nil {
		return a.theme.Error.Render("scan failed: " + a.scanErr.Error())
	}
	if len(a.targets) == 0 {
		return a.theme.Success.Render("Nothing to purge — every project is clean.")
	}

	var totalSelected, totalAll int64
	var lines []string
	for i, t := range a.targets {
		totalAll += t.Size
		check := a.theme.Dim.Render("☐")
		if a.selected[t.Path] {
			check = a.theme.Success.Render("☑")
			totalSelected += t.Size
		}
		marker := "  "
		if i == a.cursor {
			marker = a.theme.Subtitle.Render("➤ ")
		}
		recent := ""
		if t.Recent {
			recent = "  " + components.Badge(a.theme, components.BadgeWarning, " Recent ")
		}
		stack := components.Badge(a.theme, components.BadgeInfo, string(t.Stack))
		size := a.theme.Strong.Render(formatBytes(t.Size))
		row := fmt.Sprintf("%s%s  %s  %-30s  %s  %s%s",
			marker, check, stack,
			truncate(t.Project, 30),
			a.theme.Subtle.Render(t.Pattern),
			size,
			recent,
		)
		lines = append(lines, row)
	}

	body := strings.Join(lines, "\n")
	header := fmt.Sprintf(" Purge — %s found / %s selected ",
		formatBytes(totalAll), formatBytes(totalSelected))

	panel := components.Panel(a.theme, header, body,
		components.PanelOpts{Width: a.width - 4, Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "space", Action: "toggle"},
		{Key: "a", Action: "all"},
		{Key: "enter", Action: "purge (2.8)"},
		{Key: "q", Action: "quit"},
	})

	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

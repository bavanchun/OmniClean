// Package purge wraps the bubbletea program that drives the
// `omniclean purge` command. Phase 2.7 implements loading + list states;
// Phase 2.8 layers the confirm, delete, and result states on top.
package purge

import (
	"context"
	"fmt"
	"os"
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
	stateError
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
		if msg.Err != nil {
			a.scanErr = msg.Err
			a.state = stateError
			return a, nil
		}
		a.targets = msg.Targets
		// Pre-select non-recent entries.
		for _, t := range msg.Targets {
			if !t.Recent {
				a.selected[t.Path] = true
			}
		}
		a.state = stateList
		return a, nil
	case batchDeleteMsg:
		for _, r := range msg.results {
			if r.Err != nil {
				a.failed[r.Target.Path] = r.Err
			} else {
				a.deleted = append(a.deleted, r.Target)
			}
		}
		a.state = stateResult
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
	case stateConfirm:
		content = a.viewConfirm()
	case stateDeleting:
		content = a.viewDeleting()
	case stateResult:
		content = a.viewResult()
	case stateError:
		content = a.viewError()
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
	key := msg.String()
	if key == "ctrl+c" {
		return a, tea.Quit
	}

	switch a.state {
	case stateList:
		return a.keyList(key)
	case stateConfirm:
		return a.keyConfirm(key)
	case stateResult:
		if key == "q" || key == "enter" {
			return a, tea.Quit
		}
	case stateError:
		if key == "q" || key == "ctrl+c" {
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *App) keyList(key string) (tea.Model, tea.Cmd) {
	if len(a.targets) == 0 && key == "q" {
		return a, tea.Quit
	}
	switch key {
	case "q":
		return a, tea.Quit
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
	case "enter":
		if len(a.selectedTargets()) == 0 {
			return a, nil
		}
		if a.cfg.NoConfirm || a.cfg.DryRun {
			a.state = stateDeleting
			return a, tea.Batch(a.spinner.Tick, a.startDelete())
		}
		a.state = stateConfirm
	}
	return a, nil
}

func (a *App) keyConfirm(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y", "enter":
		a.state = stateDeleting
		return a, tea.Batch(a.spinner.Tick, a.startDelete())
	case "n", "N", "esc", "q":
		a.state = stateList
	}
	return a, nil
}

// selectedTargets returns the user-selected targets in display order.
func (a *App) selectedTargets() []corepurge.Target {
	out := make([]corepurge.Target, 0, len(a.targets))
	for _, t := range a.targets {
		if a.selected[t.Path] {
			out = append(out, t)
		}
	}
	return out
}

// startDelete runs deletes sequentially in a goroutine. Sequential
// deletion keeps disk and TUI behaviour predictable; throughput is
// rarely the bottleneck for project artifact removal.
func (a *App) startDelete() tea.Cmd {
	targets := a.selectedTargets()
	dryRun := a.cfg.DryRun
	return func() tea.Msg {
		// We deliver each delete as its own message via a Cmd; for
		// simplicity here we synchronously run them and return a
		// single allDeletedMsg, recording per-target outcomes on the
		// returned slice. The model accumulates them via
		// applyDeleteResults below.
		var msgs []deleteDoneMsg
		for _, t := range targets {
			var err error
			if !dryRun {
				err = os.RemoveAll(t.Path)
			}
			msgs = append(msgs, deleteDoneMsg{Target: t, Err: err})
		}
		return batchDeleteMsg{results: msgs}
	}
}

// batchDeleteMsg wraps every delete result so the UI updates atomically
// when the goroutine returns. A streaming cmd-per-delete would feel
// nicer for huge selections; this can be promoted later if needed.
type batchDeleteMsg struct{ results []deleteDoneMsg }

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
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
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
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "space", Action: "toggle"},
		{Key: "a", Action: "all"},
		{Key: "enter", Action: "purge selected"},
		{Key: "q", Action: "quit"},
	})

	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func (a *App) viewConfirm() string {
	selected := a.selectedTargets()
	var total int64
	for _, t := range selected {
		total += t.Size
	}
	header := a.theme.Strong.Render(fmt.Sprintf("Delete %d directories totalling %s?", len(selected), formatBytes(total)))
	warn := components.Badge(a.theme, components.BadgeWarning, " Permanent — files will be removed ")
	body := lipgloss.JoinVertical(lipgloss.Left,
		header, "", warn, "",
		a.theme.Body.Render("y / enter   confirm and delete"),
		a.theme.Body.Render("n / esc     back to selection"),
	)
	panel := components.Panel(a.theme, " Confirm purge ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "y", Action: "delete"},
		{Key: "n", Action: "back"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func (a *App) viewDeleting() string {
	body := fmt.Sprintf("  %s  %s",
		a.spinner.View(),
		a.theme.Body.Render("Removing selected artifacts…"),
	)
	panel := components.Panel(a.theme, " Purging ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	return panel
}

func (a *App) viewResult() string {
	var freed int64
	for _, t := range a.deleted {
		freed += t.Size
	}
	dryNote := ""
	if a.cfg.DryRun {
		dryNote = " " + components.Badge(a.theme, components.BadgeDryRun, " DRY RUN ")
	}
	chips := []string{
		components.Badge(a.theme, components.BadgeSuccess, fmt.Sprintf(" %d removed ", len(a.deleted))),
	}
	if len(a.failed) > 0 {
		chips = append(chips, components.Badge(a.theme, components.BadgeError, fmt.Sprintf(" %d failed ", len(a.failed))))
	}
	chips = append(chips, components.Badge(a.theme, components.BadgeInfo, fmt.Sprintf(" %s freed ", formatBytes(freed))))
	header := lipgloss.JoinHorizontal(lipgloss.Left, chips...) + dryNote

	var rows []string
	for _, t := range a.deleted {
		rows = append(rows, "  "+a.theme.Success.Render("✓")+"  "+a.theme.Strong.Render(t.Path)+"  "+a.theme.Subtle.Render(formatBytes(t.Size)))
	}
	for path, err := range a.failed {
		rows = append(rows, "  "+a.theme.Error.Render("✗")+"  "+a.theme.Strong.Render(path)+"  "+a.theme.Error.Render(err.Error()))
	}
	body := strings.Join(rows, "\n")
	panel := components.Panel(a.theme, " Purge results ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, header, "", panel, "", footer)
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

// safeWidth returns the panel content width, ensuring a minimum of 40.
func (a *App) safeWidth() int {
	w := a.safeWidth()
	if w < 40 {
		return 40
	}
	return w
}

// viewError renders the error state shown when scanning fails.
func (a *App) viewError() string {
	body := a.theme.Error.Render("Scan failed: " + a.scanErr.Error())
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "q", Action: "quit"},
	})
	panel := components.Panel(a.theme, " Error ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

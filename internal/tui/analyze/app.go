// Package analyze drives the interactive disk explorer used by
// `omniclean analyze`. Phase 3.7 implements loading + list views with
// the bar renderer; navigation/large-files/trash land in subsequent
// commits.
package analyze

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	coreanalyze "github.com/bavanchun/OmniClean/internal/analyze"
	"github.com/bavanchun/OmniClean/internal/tui/components"
	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// Config wires the cobra command to the TUI program.
type Config struct {
	Path    string
	Options coreanalyze.Options
}

type viewState int

const (
	stateLoading viewState = iota
	stateList
	stateError
)

// scanDoneMsg is delivered when the disk scan finishes.
type scanDoneMsg struct {
	Result coreanalyze.Result
	Err    error
}

// App is the root Bubbletea model.
type App struct {
	cfg     Config
	state   viewState
	theme   theme.Styles
	spinner spinner.Model

	width, height int

	result       coreanalyze.Result
	err          error
	cursor       int
	largeCursor  int
	showLarge    bool
	history      *coreanalyze.History
}

// New constructs the App from the supplied Config.
func New(cfg Config) *App {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary))
	return &App{
		cfg:     cfg,
		state:   stateLoading,
		theme:   theme.New(),
		spinner: sp,
		history: coreanalyze.NewHistory(0),
	}
}

// Run starts the TUI program.
func (a *App) Run(ctx context.Context) error {
	p := tea.NewProgram(a, tea.WithContext(ctx))
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("run analyze tui: %w", err)
	}
	return nil
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.spinner.Tick, a.startScan(a.cfg.Path))
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "L":
			a.showLarge = !a.showLarge
			a.largeCursor = 0
		}
		if a.showLarge {
			return a.handleLargeKey(msg)
		}
		switch msg.String() {
		case "up", "k":
			if a.cursor > 0 {
				a.cursor--
			}
		case "down", "j":
			if a.cursor < len(a.result.Entries)-1 {
				a.cursor++
			}
		case "enter", "right", "l":
			return a.openSelected()
		case "esc", "left", "h", "backspace":
			return a.goBack()
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	case scanDoneMsg:
		if msg.Err != nil {
			a.err = msg.Err
			a.state = stateError
			return a, nil
		}
		a.result = msg.Result
		a.state = stateList
	}
	return a, nil
}

func (a *App) View() tea.View {
	var content string
	switch a.state {
	case stateLoading:
		content = a.viewLoading()
	case stateError:
		content = a.theme.Error.Render("Scan failed: " + a.err.Error())
	default:
		if a.showLarge {
			content = a.viewLargeFiles()
		} else {
			content = a.viewList()
		}
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// handleLargeKey routes key presses while the large-files overlay is
// active. Esc/L closes it; up/down move the cursor.
func (a *App) handleLargeKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "L":
		a.showLarge = false
	case "up", "k":
		if a.largeCursor > 0 {
			a.largeCursor--
		}
	case "down", "j":
		if a.largeCursor < len(a.result.LargeFiles)-1 {
			a.largeCursor++
		}
	}
	return a, nil
}

// openSelected pushes the current view onto history and triggers a
// new scan of the highlighted directory. No-op when the cursor is on a
// file or the result list is empty.
func (a *App) openSelected() (tea.Model, tea.Cmd) {
	if a.state != stateList || len(a.result.Entries) == 0 {
		return a, nil
	}
	sel := a.result.Entries[a.cursor]
	if !sel.IsDir {
		return a, nil
	}
	a.history.Push(coreanalyze.HistoryEntry{Result: a.result, Cursor: a.cursor})
	a.cfg.Path = sel.Path
	a.state = stateLoading
	a.cursor = 0
	return a, tea.Batch(a.spinner.Tick, a.startScan(sel.Path))
}

// goBack pops the last view off history; when history is empty we
// instead exit the program so users get a single clear path out.
func (a *App) goBack() (tea.Model, tea.Cmd) {
	prev, ok := a.history.Pop()
	if !ok {
		return a, tea.Quit
	}
	a.result = prev.Result
	a.cursor = prev.Cursor
	a.cfg.Path = prev.Result.Path
	a.state = stateList
	return a, nil
}

func (a *App) startScan(path string) tea.Cmd {
	opts := a.cfg.Options
	return func() tea.Msg {
		res, err := coreanalyze.NewWalker().Scan(context.Background(), path, opts)
		return scanDoneMsg{Result: res, Err: err}
	}
}

func (a *App) viewLoading() string {
	body := fmt.Sprintf("  %s  %s",
		a.spinner.View(),
		a.theme.Body.Render("Scanning "+a.cfg.Path),
	)
	panel := components.Panel(a.theme, " Disk explorer ", body,
		components.PanelOpts{Width: a.width - 4, Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{{Key: "ctrl+c", Action: "cancel"}})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func (a *App) viewList() string {
	crumbs := strings.Repeat("◂ ", a.history.Len())
	header := fmt.Sprintf(" %s%s   Total: %s   Files: %d ",
		crumbs, a.cfg.Path, formatBytes(a.result.TotalSize), a.result.TotalFiles)

	var lines []string
	total := a.result.TotalSize
	if total == 0 {
		total = 1
	}
	for i, e := range a.result.Entries {
		pct := float64(e.Size) / float64(total)
		marker := "  "
		if i == a.cursor {
			marker = a.theme.Subtitle.Render("➤ ")
		}
		icon := "📄"
		if e.IsDir {
			icon = "📁"
		}
		row := fmt.Sprintf("%s%s  %5.1f%%  %s  %-30s  %s",
			marker,
			a.theme.Subtitle.Render(bar(20, pct)),
			pct*100,
			icon,
			truncate(e.Name, 30),
			a.theme.Strong.Render(formatBytes(e.Size)),
		)
		lines = append(lines, row)
	}
	body := strings.Join(lines, "\n")
	panel := components.Panel(a.theme, header, body,
		components.PanelOpts{Width: a.width - 4, Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "enter", Action: "open"},
		{Key: "esc", Action: "back"},
		{Key: "L", Action: "large files (3.9)"},
		{Key: "del", Action: "trash (3.10)"},
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func (a *App) viewLargeFiles() string {
	if len(a.result.LargeFiles) == 0 {
		body := a.theme.Subtle.Render("  No files above the size threshold in this tree.")
		panel := components.Panel(a.theme, " Large files ", body,
			components.PanelOpts{Width: a.width - 4, Accent: true})
		footer := components.KeyHints(a.theme, []components.KeyHint{
			{Key: "L / esc", Action: "close"},
			{Key: "q", Action: "quit"},
		})
		return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
	}
	var lines []string
	for i, f := range a.result.LargeFiles {
		marker := "  "
		if i == a.largeCursor {
			marker = a.theme.Subtitle.Render("➤ ")
		}
		row := fmt.Sprintf("%s%s   %-40s   %s",
			marker,
			a.theme.Strong.Render(formatBytes(f.Size)),
			truncate(f.Name, 40),
			a.theme.Subtle.Render(f.Path),
		)
		lines = append(lines, row)
	}
	body := strings.Join(lines, "\n")
	panel := components.Panel(a.theme, fmt.Sprintf(" Large files — top %d ", len(a.result.LargeFiles)),
		body, components.PanelOpts{Width: a.width - 4, Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "del", Action: "trash (3.10)"},
		{Key: "L / esc", Action: "close"},
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

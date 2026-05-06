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
	stateConfirmTrash
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

	result           coreanalyze.Result
	err              error
	cursor           int
	scrollOffset     int // viewport offset for the entry list
	largeCursor      int
	largeScrollOffset int // viewport offset for the large-files overlay
	showLarge        bool
	history          *coreanalyze.History
	pendingPath      string // path queued for trash when in stateConfirmTrash
	pendingFile      bool   // true when pending target is a single file (large-files overlay)
	statusMsg        string // transient status (e.g. "trashed: foo")
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

// safeWidth returns a content width that is safe even before the first
// WindowSizeMsg arrives (a.width == 0 → default 40 min).
func (a *App) safeWidth() int {
	w := a.width - 4
	if w < 40 {
		return 40
	}
	return w
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
		if a.state == stateConfirmTrash {
			return a.handleConfirmKey(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "L":
			a.showLarge = !a.showLarge
			a.largeCursor = 0
			a.largeScrollOffset = 0
		case "d", "delete":
			return a.queueTrash()
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
	case trashDoneMsg:
		if msg.Err != nil {
			a.statusMsg = "trash failed: " + msg.Err.Error()
			a.state = stateList
			return a, nil
		}
		a.statusMsg = "trashed: " + msg.Path
		a.state = stateLoading
		return a, tea.Batch(a.spinner.Tick, a.startScan(a.cfg.Path))
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
	case stateConfirmTrash:
		content = a.viewConfirmTrash()
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
		a.largeScrollOffset = 0
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
	a.scrollOffset = 0
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
	a.scrollOffset = 0
	a.cfg.Path = prev.Result.Path
	a.state = stateList
	return a, nil
}

// queueTrash captures the highlighted entry (or large file) and moves
// to a confirmation state. We never trash silently.
func (a *App) queueTrash() (tea.Model, tea.Cmd) {
	if a.showLarge {
		if len(a.result.LargeFiles) == 0 {
			return a, nil
		}
		f := a.result.LargeFiles[a.largeCursor]
		a.pendingPath = f.Path
		a.pendingFile = true
	} else {
		if len(a.result.Entries) == 0 {
			return a, nil
		}
		e := a.result.Entries[a.cursor]
		a.pendingPath = e.Path
		a.pendingFile = !e.IsDir
	}
	a.state = stateConfirmTrash
	return a, nil
}

// trashDoneMsg signals MoveToTrash finished for a single path.
type trashDoneMsg struct {
	Path string
	Err  error
}

func (a *App) handleConfirmKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		path := a.pendingPath
		a.state = stateList
		return a, func() tea.Msg {
			return trashDoneMsg{Path: path, Err: coreanalyze.MoveToTrash(path)}
		}
	case "n", "N", "esc", "q":
		a.state = stateList
	}
	return a, nil
}

func (a *App) viewConfirmTrash() string {
	kind := "directory"
	if a.pendingFile {
		kind = "file"
	}
	header := a.theme.Strong.Render(fmt.Sprintf("Move %s to trash?", kind))
	body := lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		a.theme.Body.Render("  "+a.pendingPath), "",
		components.Badge(a.theme, components.BadgeWarning, " Files will go to your OS Trash / Recycle Bin "),
		"",
		a.theme.Body.Render("y / enter   confirm"),
		a.theme.Body.Render("n / esc     cancel"),
	)
	panel := components.Panel(a.theme, " Confirm trash ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	return panel
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
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{{Key: "ctrl+c", Action: "cancel"}})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func (a *App) viewList() string {
	total := len(a.result.Entries)
	visibleRows := a.height - 8
	if visibleRows < 5 {
		visibleRows = 5
	}

	// Clamp scroll offset so cursor stays in view.
	if a.cursor >= a.scrollOffset+visibleRows {
		a.scrollOffset = a.cursor - visibleRows + 1
	}
	if a.cursor < a.scrollOffset {
		a.scrollOffset = a.cursor
	}

	crumbs := strings.Repeat("◂ ", a.history.Len())
	header := fmt.Sprintf(" %s%s   Total: %s   Files: %d ",
		crumbs, a.cfg.Path, formatBytes(a.result.TotalSize), a.result.TotalFiles)
	if a.statusMsg != "" {
		header += " · " + a.statusMsg
	}
	if total > 0 {
		end := a.scrollOffset + visibleRows
		if end > total {
			end = total
		}
		header += fmt.Sprintf(" [%d-%d/%d]", a.scrollOffset+1, end, total)
	}

	var lines []string
	totalSize := a.result.TotalSize
	if totalSize == 0 {
		totalSize = 1
	}
	end := a.scrollOffset + visibleRows
	if end > total {
		end = total
	}
	for i, e := range a.result.Entries[a.scrollOffset:end] {
		idx := a.scrollOffset + i
		pct := float64(e.Size) / float64(totalSize)
		marker := "  "
		if idx == a.cursor {
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
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "enter", Action: "open"},
		{Key: "esc", Action: "back"},
		{Key: "L", Action: "large files"},
		{Key: "d", Action: "trash"},
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

func (a *App) viewLargeFiles() string {
	if len(a.result.LargeFiles) == 0 {
		body := a.theme.Subtle.Render("  No files above the size threshold in this tree.")
		panel := components.Panel(a.theme, " Large files ", body,
			components.PanelOpts{Width: a.safeWidth(), Accent: true})
		footer := components.KeyHints(a.theme, []components.KeyHint{
			{Key: "L / esc", Action: "close"},
			{Key: "q", Action: "quit"},
		})
		return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
	}

	total := len(a.result.LargeFiles)
	visibleRows := a.height - 8
	if visibleRows < 5 {
		visibleRows = 5
	}

	// Clamp scroll offset so largeCursor stays in view.
	if a.largeCursor >= a.largeScrollOffset+visibleRows {
		a.largeScrollOffset = a.largeCursor - visibleRows + 1
	}
	if a.largeCursor < a.largeScrollOffset {
		a.largeScrollOffset = a.largeCursor
	}

	end := a.largeScrollOffset + visibleRows
	if end > total {
		end = total
	}

	var lines []string
	for i, f := range a.result.LargeFiles[a.largeScrollOffset:end] {
		idx := a.largeScrollOffset + i
		marker := "  "
		if idx == a.largeCursor {
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
	header := fmt.Sprintf(" Large files — top %d  [%d-%d/%d] ", total, a.largeScrollOffset+1, end, total)
	panel := components.Panel(a.theme, header, body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "d", Action: "trash"},
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

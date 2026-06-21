// Package cleanup provides the self-contained "Cleanup Suggestions" TUI. It
// lists removable orphan/leaf packages (aggregated by internal/cleanup) and
// runs a thin select → confirm → delete flow. It deliberately does NOT import
// the monolithic uninstall App's unexported list/confirm models; following the
// internal/tui/appuninstall precedent, it reimplements a minimal state machine.
package cleanup

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/cleaner"
	agg "github.com/bavanchun/OmniClean/internal/cleanup"
	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// viewState is the TUI finite-state-machine discriminant.
type viewState int

const (
	stateLoading  viewState = iota // aggregating candidates
	stateList                      // selecting candidates to remove
	stateConfirm                   // confirm removal
	stateDeleting                  // removal in progress (normal + sudo)
	stateResult                    // summary
)

// Config carries caller options into the cleanup TUI.
type Config struct {
	Detectors []detector.Detector
	DryRun    bool
}

// App is the root Bubbletea model for the cleanup TUI.
type App struct {
	cfg     Config
	state   viewState
	theme   theme.Styles
	sp      spinner.Model
	keys    keyMap
	cleaner *cleaner.Cleaner
	detMap  map[string]detector.Detector

	width, height int
	scrollOffset  int

	candidates []pkg.Package
	selected   map[string]bool // key(p) -> selected
	cursor     int

	// delete bookkeeping
	results              []pkg.UninstallResult
	sudoQueue            []pkg.Package
	normalDone, sudoDone bool
}

// New constructs the cleanup App with sensible defaults.
func New(cfg Config) *App {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary))

	detMap := make(map[string]detector.Detector, len(cfg.Detectors))
	for _, d := range cfg.Detectors {
		detMap[d.Name()] = d
	}

	return &App{
		cfg:      cfg,
		state:    stateLoading,
		theme:    theme.New(),
		sp:       sp,
		keys:     defaultKeys(),
		cleaner:  cleaner.New(cfg.Detectors),
		detMap:   detMap,
		selected: map[string]bool{},
	}
}

// Run starts the blocking Bubbletea event loop.
func (a *App) Run(ctx context.Context) error {
	p := tea.NewProgram(a, tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run cleanup tui: %w", err)
	}
	return nil
}

// key uniquely identifies a candidate package across managers.
func key(p pkg.Package) string { return string(p.Manager) + ":" + p.Name }

// ── Bubbletea contract ──────────────────────────────────────────────────────

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.sp.Tick, a.startLoad())
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.sp, cmd = a.sp.Update(msg)
		return a, cmd

	case tea.KeyPressMsg:
		return a.handleKey(msg.String())

	case loadDoneMsg:
		a.candidates = msg.candidates
		a.state = stateList

	case normalDeleteDoneMsg:
		a.results = append(a.results, msg.results...)
		a.normalDone = true
		return a.maybeFinishDelete()

	case sudoDeleteDoneMsg:
		a.results = append(a.results, msg.result)
		return a, a.execNextSudo()

	case sudoAllDoneMsg:
		a.sudoDone = true
		return a.maybeFinishDelete()
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
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// ── Key handling ────────────────────────────────────────────────────────────

func (a *App) handleKey(k string) (tea.Model, tea.Cmd) {
	if a.keys.isQuit(k) {
		return a, tea.Quit
	}
	switch a.state {
	case stateList:
		return a.keyList(k)
	case stateConfirm:
		return a.keyConfirm(k)
	}
	return a, nil
}

func (a *App) keyList(k string) (tea.Model, tea.Cmd) {
	switch {
	case a.keys.isUp(k):
		if a.cursor > 0 {
			a.cursor--
		}
		a.clampScroll()
	case a.keys.isDown(k):
		if a.cursor < len(a.candidates)-1 {
			a.cursor++
		}
		a.clampScroll()
	case k == a.keys.Space:
		if len(a.candidates) > 0 {
			p := a.candidates[a.cursor]
			a.selected[key(p)] = !a.selected[key(p)]
		}
	case k == a.keys.All:
		allOn := true
		for _, p := range a.candidates {
			if !a.selected[key(p)] {
				allOn = false
				break
			}
		}
		for _, p := range a.candidates {
			a.selected[key(p)] = !allOn
		}
	case k == a.keys.Enter:
		if len(a.selectedPkgs()) > 0 {
			a.state = stateConfirm
		}
	}
	return a, nil
}

func (a *App) keyConfirm(k string) (tea.Model, tea.Cmd) {
	switch k {
	case a.keys.Yes, a.keys.Enter:
		a.state = stateDeleting
		return a, tea.Batch(a.sp.Tick, a.startDelete())
	case a.keys.No, a.keys.Back:
		a.state = stateList
	}
	return a, nil
}

// ── Commands ────────────────────────────────────────────────────────────────

func (a *App) startLoad() tea.Cmd {
	dets := a.cfg.Detectors
	return func() tea.Msg {
		return loadDoneMsg{candidates: agg.Aggregate(context.Background(), dets)}
	}
}

// startDelete partitions the selection into sudo and normal packages, then
// kicks off both flows. Sudo removals sequence one-at-a-time via tea.Exec so
// each can prompt for a password (mirrors the uninstall App precedent); normal
// removals run in the background through the shared cleaner.
func (a *App) startDelete() tea.Cmd {
	sel := a.selectedPkgs()
	a.results = nil
	a.normalDone, a.sudoDone = false, false

	var normal, sudo []pkg.Package
	for _, p := range sel {
		d, ok := a.detMap[string(p.Manager)]
		if ok && d.NeedsSudo() && !a.cfg.DryRun {
			sudo = append(sudo, p)
		} else {
			normal = append(normal, p)
		}
	}
	a.sudoQueue = sudo

	var cmds []tea.Cmd
	if len(normal) > 0 {
		cmds = append(cmds, a.normalDeleteCmd(normal))
	} else {
		a.normalDone = true
	}
	if len(sudo) > 0 {
		cmds = append(cmds, a.execNextSudo())
	} else {
		a.sudoDone = true
	}
	return tea.Batch(cmds...)
}

func (a *App) normalDeleteCmd(packages []pkg.Package) tea.Cmd {
	cl := a.cleaner
	dryRun := a.cfg.DryRun
	return func() tea.Msg {
		return normalDeleteDoneMsg{results: cl.Uninstall(context.Background(), packages, dryRun)}
	}
}

// execNextSudo pops the next sudo package and runs it via tea.Exec, or signals
// completion when the queue is empty.
func (a *App) execNextSudo() tea.Cmd {
	if len(a.sudoQueue) == 0 {
		return func() tea.Msg { return sudoAllDoneMsg{} }
	}
	p := a.sudoQueue[0]
	a.sudoQueue = a.sudoQueue[1:]

	d, ok := a.detMap[string(p.Manager)]
	if !ok {
		r := pkg.UninstallResult{Package: p, Err: fmt.Errorf("no detector for %q", p.Manager)}
		return func() tea.Msg { return sudoDeleteDoneMsg{result: r} }
	}
	execCmd := d.UninstallExecCmd(p)
	if execCmd == nil {
		r := pkg.UninstallResult{Package: p, Err: fmt.Errorf("detector %q returned nil exec cmd", d.Name())}
		return func() tea.Msg { return sudoDeleteDoneMsg{result: r} }
	}
	captured := p
	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return sudoDeleteDoneMsg{result: pkg.UninstallResult{Package: captured, Err: err}}
	})
}

// maybeFinishDelete transitions to the result screen once both flows finish.
func (a *App) maybeFinishDelete() (tea.Model, tea.Cmd) {
	if a.normalDone && a.sudoDone {
		a.state = stateResult
	}
	return a, nil
}

// ── Helpers ─────────────────────────────────────────────────────────────────

// selectedPkgs returns the toggled-on candidates in display order.
func (a *App) selectedPkgs() []pkg.Package {
	out := make([]pkg.Package, 0, len(a.candidates))
	for _, p := range a.candidates {
		if a.selected[key(p)] {
			out = append(out, p)
		}
	}
	return out
}

func (a *App) safeWidth() int {
	w := a.width - 4
	if w < 40 {
		return 40
	}
	return w
}

func (a *App) visibleRows() int {
	vr := a.height - 8
	if vr < 5 {
		return 5
	}
	return vr
}

func (a *App) clampScroll() {
	vr := a.visibleRows()
	if a.cursor >= a.scrollOffset+vr {
		a.scrollOffset = a.cursor - vr + 1
	}
	if a.cursor < a.scrollOffset {
		a.scrollOffset = a.cursor
	}
	maxOffset := len(a.candidates) - vr
	if maxOffset < 0 {
		maxOffset = 0
	}
	if a.scrollOffset > maxOffset {
		a.scrollOffset = maxOffset
	}
}

// roleBadge labels a package's removable role: orphan vs leaf.
func roleBadge(r pkg.Role) string {
	if r == pkg.RoleOrphan {
		return "orphan"
	}
	return "leaf"
}

// formatBytes renders a byte count compactly.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// truncate shortens s to at most n runes with an ellipsis.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

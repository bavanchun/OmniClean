//go:build darwin

// Package appuninstall provides the Bubbletea TUI for the `omniclean uninstall`
// command. It scans .app bundles, lets the user select apps to remove, and
// optionally cleans up leftover preference / cache / support files.
package appuninstall

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appuninstall "github.com/bavanchun/OmniClean/internal/appuninstall"
	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// viewState is the TUI's finite state machine discriminant.
type viewState int

const (
	stateLoading           viewState = iota // scanning .app bundles
	stateList                               // selecting apps to remove
	stateDetail                             // detail view for one app
	stateConfirmBundle                      // confirm deletion of selected bundles
	stateDeletingBundle                     // bundle removal in progress
	stateLeftoverScan                       // scanning for orphan files
	stateConfirmLeftovers                   // confirm leftover cleanup
	stateDeletingLeftovers                  // leftover removal in progress
	stateResult                             // summary screen
	stateError                              // unrecoverable error
)

// Config carries all caller-supplied options into the TUI.
type Config struct {
	// Roots overrides the scan directories. nil = use appuninstall.ScanRoots().
	Roots []string
	// DryRun skips actual filesystem mutations but runs every other code path.
	DryRun bool
}

// App is the root Bubbletea model for the appuninstall TUI.
type App struct {
	cfg   Config
	state viewState
	theme theme.Styles
	sp    spinner.Model
	keys  keyMap

	width, height int
	scrollOffset  int

	// scan results
	bundles  []appuninstall.Bundle
	selected map[string]bool // bundle.Path -> selected
	cursor   int
	scanErr  error

	// detail view
	detailBundle    appuninstall.Bundle
	detailLeftovers []appuninstall.LeftoverEntry // populated lazily by viewDetail

	// post-bundle-delete state
	deletedBundles []appuninstall.Bundle
	bundleResults  []appuninstall.DeleteResult

	// post-leftover state
	leftovers       []appuninstall.LeftoverEntry
	leftoverResults []appuninstall.DeleteResult
}

// New constructs the App from cfg and wires sensible defaults.
func New(cfg Config) *App {
	if cfg.Roots == nil {
		cfg.Roots = appuninstall.ScanRoots()
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary))

	return &App{
		cfg:      cfg,
		state:    stateLoading,
		theme:    theme.New(),
		sp:       sp,
		keys:     defaultKeys(),
		selected: map[string]bool{},
	}
}

// Run starts the blocking Bubbletea event loop.
func (a *App) Run(ctx context.Context) error {
	p := tea.NewProgram(a, tea.WithContext(ctx))
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("run appuninstall tui: %w", err)
	}
	return nil
}

// ── Bubbletea contract ────────────────────────────────────────────────────────

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.sp.Tick, a.startScan())
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

	case scanDoneMsg:
		if msg.err != nil {
			a.scanErr = msg.err
			a.state = stateError
			return a, nil
		}
		a.bundles = msg.bundles
		a.state = stateList

	case bundleDeleteDoneMsg:
		a.bundleResults = msg.results
		// Collect bundles that were actually deleted (no error).
		for _, r := range msg.results {
			if r.Err == nil {
				for _, b := range a.bundles {
					if b.Path == r.Path {
						a.deletedBundles = append(a.deletedBundles, b)
						break
					}
				}
			}
		}
		a.state = stateLeftoverScan
		return a, tea.Batch(a.sp.Tick, a.startLeftoverScan())

	case leftoverScanDoneMsg:
		if msg.err != nil {
			// Non-fatal: skip leftovers and go straight to results.
			a.state = stateResult
			return a, nil
		}
		a.leftovers = msg.entries
		if len(msg.entries) == 0 {
			a.state = stateResult
			return a, nil
		}
		a.state = stateConfirmLeftovers

	case leftoverDeleteDoneMsg:
		a.leftoverResults = msg.results
		a.state = stateResult
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
	case stateDetail:
		content = a.viewDetail()
	case stateConfirmBundle:
		content = a.viewConfirmBundle()
	case stateDeletingBundle:
		content = a.viewDeletingBundle()
	case stateLeftoverScan:
		content = a.viewLeftoverScan()
	case stateConfirmLeftovers:
		content = a.viewConfirmLeftovers()
	case stateDeletingLeftovers:
		content = a.viewDeletingLeftovers()
	case stateResult:
		content = a.viewResult()
	case stateError:
		content = a.viewError()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// ── Key handling ──────────────────────────────────────────────────────────────

func (a *App) handleKey(key string) (tea.Model, tea.Cmd) {
	if a.keys.isQuit(key) {
		return a, tea.Quit
	}

	switch a.state {
	case stateList:
		return a.keyList(key)
	case stateDetail:
		return a.keyDetail(key)
	case stateConfirmBundle:
		return a.keyConfirmBundle(key)
	case stateConfirmLeftovers:
		return a.keyConfirmLeftovers(key)
	case stateResult, stateError:
		// quit already handled above
	}
	return a, nil
}

func (a *App) keyList(key string) (tea.Model, tea.Cmd) {
	switch {
	case a.keys.isUp(key):
		if a.cursor > 0 {
			a.cursor--
		}
		a.clampScroll(len(a.bundles))

	case a.keys.isDown(key):
		if a.cursor < len(a.bundles)-1 {
			a.cursor++
		}
		a.clampScroll(len(a.bundles))

	case key == a.keys.Space:
		if len(a.bundles) > 0 {
			b := a.bundles[a.cursor]
			a.selected[b.Path] = !a.selected[b.Path]
		}

	case key == a.keys.All:
		allOn := true
		for _, b := range a.bundles {
			if !a.selected[b.Path] {
				allOn = false
				break
			}
		}
		for _, b := range a.bundles {
			a.selected[b.Path] = !allOn
		}

	case key == a.keys.Detail:
		if len(a.bundles) > 0 {
			a.detailBundle = a.bundles[a.cursor]
			a.detailLeftovers = nil
			a.state = stateDetail
		}

	case key == a.keys.Enter:
		if len(a.selectedBundles()) == 0 {
			return a, nil
		}
		a.state = stateConfirmBundle
	}
	return a, nil
}

func (a *App) keyDetail(key string) (tea.Model, tea.Cmd) {
	switch key {
	case a.keys.Back:
		a.state = stateList
	case a.keys.Space:
		b := a.detailBundle
		a.selected[b.Path] = !a.selected[b.Path]
	}
	return a, nil
}

func (a *App) keyConfirmBundle(key string) (tea.Model, tea.Cmd) {
	switch key {
	case a.keys.Yes, a.keys.Enter:
		a.state = stateDeletingBundle
		return a, tea.Batch(a.sp.Tick, a.startDeleteBundles())
	case a.keys.No, a.keys.Back:
		a.state = stateList
	}
	return a, nil
}

func (a *App) keyConfirmLeftovers(key string) (tea.Model, tea.Cmd) {
	switch key {
	case a.keys.Yes, a.keys.Enter:
		a.state = stateDeletingLeftovers
		return a, tea.Batch(a.sp.Tick, a.startDeleteLeftovers())
	case a.keys.No, a.keys.Back:
		// User chose to skip leftover cleanup.
		a.state = stateResult
	}
	return a, nil
}

// ── Commands ──────────────────────────────────────────────────────────────────

func (a *App) startScan() tea.Cmd {
	roots := a.cfg.Roots
	return func() tea.Msg {
		bundles, err := appuninstall.Scan(context.Background(), roots)
		return scanDoneMsg{bundles: bundles, err: err}
	}
}

func (a *App) startDeleteBundles() tea.Cmd {
	targets := a.selectedBundles()
	dryRun := a.cfg.DryRun
	return func() tea.Msg {
		results := make([]appuninstall.DeleteResult, 0, len(targets))
		ctx := context.Background()
		for _, b := range targets {
			results = append(results, appuninstall.DeleteBundle(ctx, b, dryRun))
		}
		return bundleDeleteDoneMsg{results: results}
	}
}

func (a *App) startLeftoverScan() tea.Cmd {
	deleted := a.deletedBundles
	return func() tea.Msg {
		ctx := context.Background()
		var all []appuninstall.LeftoverEntry
		for _, b := range deleted {
			entries, err := appuninstall.FindLeftovers(ctx, b)
			if err != nil {
				return leftoverScanDoneMsg{err: err}
			}
			all = append(all, entries...)
		}
		return leftoverScanDoneMsg{entries: all}
	}
}

func (a *App) startDeleteLeftovers() tea.Cmd {
	entries := a.leftovers
	dryRun := a.cfg.DryRun
	return func() tea.Msg {
		results := appuninstall.DeleteLeftovers(context.Background(), entries, dryRun)
		return leftoverDeleteDoneMsg{results: results}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// selectedBundles returns bundles the user has toggled on, preserving display order.
func (a *App) selectedBundles() []appuninstall.Bundle {
	out := make([]appuninstall.Bundle, 0, len(a.bundles))
	for _, b := range a.bundles {
		if a.selected[b.Path] {
			out = append(out, b)
		}
	}
	return out
}

// safeWidth returns the panel content width, ensuring a minimum of 40 columns
// even before the first WindowSizeMsg arrives (a.width == 0).
func (a *App) safeWidth() int {
	w := a.width - 4
	if w < 40 {
		return 40
	}
	return w
}

// formatBytes renders a byte count as a compact human-readable string.
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

// truncate shortens s to at most n runes, appending "…" when truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}

// totalSize sums the Size field of a slice of bundles.
func totalBundleSize(bundles []appuninstall.Bundle) int64 {
	var t int64
	for _, b := range bundles {
		t += b.Size
	}
	return t
}

// totalLeftoverSize sums the Size field of a slice of leftover entries.
func totalLeftoverSize(entries []appuninstall.LeftoverEntry) int64 {
	var t int64
	for _, e := range entries {
		t += e.Size
	}
	return t
}

// visibleRows computes the number of list rows that fit in the current terminal.
func (a *App) visibleRows() int {
	vr := a.height - 8
	if vr < 5 {
		return 5
	}
	return vr
}

// clampScroll adjusts a.scrollOffset so a.cursor stays inside the viewport.
func (a *App) clampScroll(total int) {
	vr := a.visibleRows()
	if a.cursor >= a.scrollOffset+vr {
		a.scrollOffset = a.cursor - vr + 1
	}
	if a.cursor < a.scrollOffset {
		a.scrollOffset = a.cursor
	}
	maxOffset := total - vr
	if maxOffset < 0 {
		maxOffset = 0
	}
	if a.scrollOffset > maxOffset {
		a.scrollOffset = maxOffset
	}
}

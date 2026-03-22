package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bavanchun/OmniClean/internal/cleaner"
	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// viewState tracks which screen is active.
type viewState int

const (
	stateLoading viewState = iota
	stateList
	stateDetail
	stateConfirm
	stateUninstalling
	stateResult
	stateError
)

// Config holds application-level configuration passed in from main.
type Config struct {
	Detectors []detector.Detector
	DryRun    bool
	NoConfirm bool
}

// App is the root Bubbletea model.
type App struct {
	config  Config
	state   viewState
	styles  Styles
	cleaner *cleaner.Cleaner

	// sub-models
	spinner spinner.Model
	list    listModel
	detail  detailModel
	confirm confirmModel

	// state data
	packages      []pkg.Package
	currentPkg    pkg.Package
	results       []pkg.UninstallResult
	err           error
	width, height int

	// sudo uninstall sequencing
	detectorMap    map[string]detector.Detector
	sudoQueue      []pkg.Package // packages waiting to be uninstalled via tea.Exec
	sudoCurrentPkg *pkg.Package  // package currently being sudo-uninstalled
	normalDone     bool          // normal (non-sudo) batch finished
	sudoDone       bool          // all sudo packages finished
}

// New creates the App model ready to be run.
func New(cfg Config) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	// Build detector map for sudo lookup
	dmap := make(map[string]detector.Detector, len(cfg.Detectors))
	for _, d := range cfg.Detectors {
		dmap[d.Name()] = d
	}

	return &App{
		config:      cfg,
		state:       stateLoading,
		styles:      DefaultStyles(),
		spinner:     s,
		cleaner:     cleaner.New(cfg.Detectors),
		detectorMap: dmap,
	}
}

// Run starts the Bubbletea program.
func (a *App) Run(ctx context.Context) error {
	p := tea.NewProgram(a, tea.WithAltScreen(), tea.WithContext(ctx))
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}
	return nil
}

// --- Bubbletea interface ---

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		a.loadPackagesCmd(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			if a.state == stateList || a.state == stateError || a.state == stateResult {
				return a, tea.Quit
			}
		case "enter", "r":
			if a.state == stateResult {
				a.state = stateLoading
				a.results = nil
				a.normalDone = false
				a.sudoDone = false
				return a, tea.Batch(a.spinner.Tick, a.loadPackagesCmd())
			}
		case "esc":
			switch a.state {
			case stateDetail:
				a.state = stateList
				return a, nil
			case stateConfirm:
				a.state = stateList
				return a, nil
			}
		}

	case PackagesLoadedMsg:
		a.packages = msg.Packages
		a.list = newListModel(a.packages, a.styles, msg.Warnings)
		a.list.list.SetSize(a.width, a.height-4)
		a.state = stateList
		return a, nil

	case PackagesLoadErrorMsg:
		a.err = msg.Err
		a.state = stateError
		return a, nil

	case confirmYesMsg:
		selected := a.list.SelectedPackages()
		if len(selected) == 0 {
			a.state = stateList
			return a, nil
		}
		a.state = stateUninstalling
		return a, a.startUninstall(selected)

	case confirmNoMsg:
		a.state = stateList
		return a, nil

	case UninstallCompleteMsg:
		// Normal (non-sudo) packages finished
		a.results = append(a.results, msg.Results...)
		a.normalDone = true
		if a.sudoDone || len(a.sudoQueue) == 0 {
			a.state = stateResult
		}
		return a, nil

	case sudoUninstallDoneMsg:
		a.sudoCurrentPkg = nil
		a.results = append(a.results, msg.result)
		return a, a.execNextSudo()

	case sudoAllDoneMsg:
		a.sudoDone = true
		if a.normalDone {
			a.state = stateResult
		}
		return a, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}

	// Delegate to sub-models
	switch a.state {
	case stateList:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				selected := a.list.SelectedPackages()
				if len(selected) > 0 {
					if a.config.NoConfirm {
						a.state = stateUninstalling
						return a, a.startUninstall(selected)
					}
					hasSudo := a.selectedNeedSudo(selected)
					a.confirm = newConfirmModel(selected, a.config.DryRun, hasSudo, a.styles, a.width)
					a.state = stateConfirm
					return a, nil
				}
			case "d":
				if item, ok := a.list.list.SelectedItem().(pkg.Package); ok {
					a.currentPkg = item
					a.detail = newDetailModel(item, a.styles)
					a.state = stateDetail
					return a, nil
				}
			}
		}
		var cmd tea.Cmd
		a.list, cmd = a.list.Update(msg)
		return a, cmd

	case stateDetail:
		var cmd tea.Cmd
		a.detail, cmd = a.detail.Update(msg)
		return a, cmd

	case stateConfirm:
		if key, ok := msg.(tea.KeyMsg); ok {
			cmd := a.confirm.HandleKey(key.String())
			if cmd != nil {
				return a, cmd
			}
		}
		var cmd tea.Cmd
		a.confirm, cmd = a.confirm.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) View() string {
	switch a.state {
	case stateLoading:
		return fmt.Sprintf("\n  %s Detecting installed packages...\n", a.spinner.View())
	case stateList:
		return a.list.View()
	case stateDetail:
		return a.detail.View()
	case stateConfirm:
		return a.confirm.View()
	case stateUninstalling:
		if a.sudoCurrentPkg != nil {
			return fmt.Sprintf("\n  Uninstalling %s (requires sudo)...\n\n  The terminal will prompt for your password.\n",
				a.sudoCurrentPkg.Name)
		}
		return fmt.Sprintf("\n  %s Uninstalling packages...\n", a.spinner.View())
	case stateResult:
		return a.resultView()
	case stateError:
		return a.styles.ErrorText.Render(fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", a.err))
	default:
		return ""
	}
}

// --- Commands ---

func (a *App) loadPackagesCmd() tea.Cmd {
	detectors := a.config.Detectors
	return func() tea.Msg {
		ctx := context.Background()
		var packages []pkg.Package
		var warnings []string
		for _, d := range detectors {
			pkgs, err := d.ListPackages(ctx)
			if err != nil {
				warnings = append(warnings, d.Name())
				continue
			}
			packages = append(packages, pkgs...)
		}
		if len(packages) == 0 && len(warnings) > 0 {
			return PackagesLoadErrorMsg{
				Err: fmt.Errorf("all detectors failed: %s", strings.Join(warnings, ", ")),
			}
		}
		return PackagesLoadedMsg{Packages: packages, Warnings: warnings}
	}
}

// startUninstall partitions packages into sudo and normal, then launches both flows.
func (a *App) startUninstall(packages []pkg.Package) tea.Cmd {
	a.results = nil
	a.normalDone = false
	a.sudoDone = false

	var sudoPkgs, normalPkgs []pkg.Package
	for _, p := range packages {
		d, ok := a.detectorMap[string(p.Manager)]
		if ok && d.NeedsSudo() && !a.config.DryRun {
			sudoPkgs = append(sudoPkgs, p)
		} else {
			normalPkgs = append(normalPkgs, p)
		}
	}

	a.sudoQueue = sudoPkgs

	var cmds []tea.Cmd

	// Normal packages run in background
	if len(normalPkgs) > 0 {
		cmds = append(cmds, a.uninstallNormalCmd(normalPkgs))
	} else {
		a.normalDone = true
	}

	// Sudo packages sequence one at a time via tea.Exec
	if len(sudoPkgs) > 0 {
		cmds = append(cmds, a.execNextSudo())
	} else {
		a.sudoDone = true
	}

	// If both are empty somehow, go directly to results
	if len(cmds) == 0 {
		a.state = stateResult
		return nil
	}

	return tea.Batch(cmds...)
}

func (a *App) uninstallNormalCmd(packages []pkg.Package) tea.Cmd {
	cl := a.cleaner
	dryRun := a.config.DryRun
	return func() tea.Msg {
		ctx := context.Background()
		results := cl.Uninstall(ctx, packages, dryRun)
		return UninstallCompleteMsg{Results: results}
	}
}

// execNextSudo pops the next sudo package and runs it via tea.Exec.
// Returns sudoAllDoneMsg when the queue is empty.
func (a *App) execNextSudo() tea.Cmd {
	if len(a.sudoQueue) == 0 {
		return func() tea.Msg { return sudoAllDoneMsg{} }
	}

	p := a.sudoQueue[0]
	a.sudoQueue = a.sudoQueue[1:]

	d, ok := a.detectorMap[string(p.Manager)]
	if !ok {
		// No detector — record error and move to next
		r := pkg.UninstallResult{
			Package: p,
			Err:     fmt.Errorf("no detector for %q", p.Manager),
		}
		return func() tea.Msg { return sudoUninstallDoneMsg{result: r} }
	}

	execCmd := d.UninstallExecCmd(p)
	if execCmd == nil {
		r := pkg.UninstallResult{
			Package: p,
			Err:     fmt.Errorf("detector %q returned nil exec cmd", d.Name()),
		}
		return func() tea.Msg { return sudoUninstallDoneMsg{result: r} }
	}

	a.sudoCurrentPkg = &p
	captured := p // capture for closure
	return tea.ExecProcess(execCmd, func(err error) tea.Msg {
		return sudoUninstallDoneMsg{result: pkg.UninstallResult{Package: captured, Err: err}}
	})
}

// selectedNeedSudo returns true if any selected package requires sudo.
func (a *App) selectedNeedSudo(packages []pkg.Package) bool {
	for _, p := range packages {
		if d, ok := a.detectorMap[string(p.Manager)]; ok && d.NeedsSudo() {
			return true
		}
	}
	return false
}

// --- Views ---

func (a *App) resultView() string {
	var b strings.Builder
	fmt.Fprintln(&b, a.styles.Title.Render("Uninstall Results"))
	fmt.Fprintln(&b)

	successCount, failedCount, dryRunCount := 0, 0, 0
	for _, r := range a.results {
		badge := a.styles.BadgeFor(string(r.Package.Manager))

		switch {
		case r.DryRunCmd != "":
			dryIcon := a.styles.DryRunBadge.Render("DRY")
			fmt.Fprintf(&b, "  %s %s %s\n", dryIcon, badge, r.Package.Name)
			fmt.Fprintf(&b, "    %s\n", a.styles.HelpBar.Render(r.DryRunCmd))
			dryRunCount++

		case r.Err != nil:
			fmt.Fprintf(&b, "  %s %s %s\n",
				a.styles.ErrorText.Render("✗"), badge, r.Package.Name)
			fmt.Fprintf(&b, "    %s\n", a.styles.ErrorText.Render(r.Err.Error()))
			failedCount++

		default:
			okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#48BB78"))
			fmt.Fprintf(&b, "  %s %s %s\n", okStyle.Render("✓"), badge, r.Package.Name)
			if len(r.Leftovers) > 0 {
				fmt.Fprintf(&b, "    %s\n",
					a.styles.HelpBar.Render("Leftover files: "+strings.Join(r.Leftovers, ", ")))
			}
			successCount++
		}
	}

	fmt.Fprintln(&b)
	var summary string
	switch {
	case dryRunCount > 0:
		summary = fmt.Sprintf("Dry run: %d commands shown", dryRunCount)
	default:
		summary = fmt.Sprintf("Done: %d removed", successCount)
		if failedCount > 0 {
			summary += fmt.Sprintf(", %d failed", failedCount)
		}
	}
	fmt.Fprintln(&b, a.styles.StatusBar.Render(summary))
	fmt.Fprint(&b, a.styles.HelpBar.Render("\n  enter/r: back to list  ·  q: quit"))

	return b.String()
}

// sudoAllDoneMsg is sent when all sudo packages have been processed.
type sudoAllDoneMsg struct{}

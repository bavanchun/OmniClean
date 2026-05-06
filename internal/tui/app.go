package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/cleaner"
	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
	"github.com/bavanchun/OmniClean/internal/tui/components"
	"github.com/bavanchun/OmniClean/internal/tui/theme"
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
	stateSettings
	stateError
)

// Config holds application-level configuration passed in from main.
type Config struct {
	Detectors []detector.Detector
	DryRun    bool
	NoConfirm bool
	Verbose   bool
}

// App is the root Bubbletea model.
type App struct {
	config  Config
	state   viewState
	styles  Styles
	theme   theme.Styles
	cleaner *cleaner.Cleaner
	keys    KeyMap

	// sub-models
	spinner  spinner.Model
	progress progressModel
	list     listModel
	detail   detailModel
	confirm  confirmModel
	settings settingsModel
	help     help.Model

	// result view
	resultTable    table.Model
	resultViewport viewport.Model

	// state data
	packages      []pkg.Package
	currentPkg    pkg.Package
	results       []pkg.UninstallResult
	err           error
	width, height int

	// loading state (per-detector progress)
	detectorCh    chan DetectorDoneMsg
	detectorNames []string
	loadPackages  []pkg.Package
	loadWarnings  []string
	doneCount     int

	// sudo uninstall sequencing
	detectorMap    map[string]detector.Detector
	sudoQueue      []pkg.Package // packages waiting to be uninstalled via tea.Exec
	sudoCurrentPkg *pkg.Package  // package currently being sudo-uninstalled
	normalDone     bool          // normal (non-sudo) batch finished
	sudoDone       bool          // all sudo packages finished

	// all available detectors (for settings — superset of config.Detectors)
	allDetectors []detector.Detector
}

// New creates the App model ready to be run.
func New(cfg Config) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary))

	styles := DefaultStyles()

	// Build detector map for sudo lookup
	dmap := make(map[string]detector.Detector, len(cfg.Detectors))
	for _, d := range cfg.Detectors {
		dmap[d.Name()] = d
	}

	keys := DefaultKeyMap()
	h := help.New()
	h.Styles = help.DefaultDarkStyles()

	return &App{
		config:       cfg,
		state:        stateLoading,
		styles:       styles,
		theme:        theme.New(),
		spinner:      s,
		progress:     newProgressModel(styles),
		cleaner:      cleaner.New(cfg.Detectors),
		detectorMap:  dmap,
		keys:         keys,
		help:         h,
		allDetectors: cfg.Detectors, // initially all available
	}
}

// Run starts the Bubbletea program.
func (a *App) Run(ctx context.Context) error {
	p := tea.NewProgram(a, tea.WithContext(ctx))
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}
	return nil
}

// --- Bubbletea interface ---

func (a *App) Init() tea.Cmd {
	return a.startLoading()
}

// startLoading resets loading state, spawns the detection goroutine, and
// returns the commands to drive the spinner and read per-detector results.
func (a *App) startLoading() tea.Cmd {
	a.detectorNames = make([]string, len(a.config.Detectors))
	for i, d := range a.config.Detectors {
		a.detectorNames[i] = d.Name()
	}
	a.detectorCh = make(chan DetectorDoneMsg)
	a.loadPackages = nil
	a.loadWarnings = nil
	a.doneCount = 0
	a.progress = newProgressModel(a.styles)

	detectors := a.config.Detectors
	ch := a.detectorCh
	go func() {
		ctx := context.Background()
		for _, d := range detectors {
			pkgs, err := d.ListPackages(ctx)
			ch <- DetectorDoneMsg{Name: d.Name(), Packages: pkgs, Err: err}
		}
		close(ch)
	}()

	return tea.Batch(a.spinner.Tick, waitForDetector(a.detectorCh), progressTick())
}

// waitForDetector returns a Cmd that reads one result from the detector channel.
func waitForDetector(ch chan DetectorDoneMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return allDetectorsDoneMsg{}
		}
		return msg
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.SetWidth(msg.Width)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, a.keys.ForceQuit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Quit):
			if a.state == stateList || a.state == stateError || a.state == stateResult {
				return a, tea.Quit
			}
		case key.Matches(msg, a.keys.Confirm) || key.Matches(msg, a.keys.Refresh):
			if a.state == stateResult {
				a.state = stateLoading
				a.results = nil
				a.normalDone = false
				a.sudoDone = false
				return a, a.startLoading()
			}
		case key.Matches(msg, a.keys.Back):
			switch a.state {
			case stateDetail:
				a.state = stateList
				return a, nil
			case stateConfirm:
				a.state = stateList
				return a, nil
			case stateSettings:
				a.state = stateList
				return a, nil
			}
		}

	case DetectorDoneMsg:
		a.doneCount++
		if msg.Err != nil {
			a.loadWarnings = append(a.loadWarnings, msg.Name)
		} else {
			a.loadPackages = append(a.loadPackages, msg.Packages...)
		}
		// Update progress target
		if len(a.detectorNames) > 0 {
			a.progress.SetTarget(float64(a.doneCount) / float64(len(a.detectorNames)))
		}
		return a, waitForDetector(a.detectorCh)

	case allDetectorsDoneMsg:
		if len(a.loadPackages) == 0 && len(a.loadWarnings) > 0 {
			a.err = fmt.Errorf("all detectors failed: %s", strings.Join(a.loadWarnings, ", "))
			a.state = stateError
			return a, nil
		}
		a.packages = a.loadPackages
		a.list = newListModel(a.packages, a.styles, a.loadWarnings)
		a.list.list.SetSize(a.width, a.height-6)
		a.state = stateList
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
			a.buildResultTable()
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
			a.buildResultTable()
		}
		return a, nil

	case settingsAppliedMsg:
		// Rebuild config with selected detectors
		newDetectors := make([]detector.Detector, 0)
		nameSet := make(map[string]bool, len(msg.SelectedManagers))
		for _, n := range msg.SelectedManagers {
			nameSet[n] = true
		}
		for _, d := range a.allDetectors {
			if nameSet[d.Name()] {
				newDetectors = append(newDetectors, d)
			}
		}
		if len(newDetectors) == 0 {
			// Don't allow empty — keep all
			newDetectors = a.allDetectors
		}
		a.config.Detectors = newDetectors
		// Rebuild detector map and cleaner
		a.detectorMap = make(map[string]detector.Detector, len(newDetectors))
		for _, d := range newDetectors {
			a.detectorMap[d.Name()] = d
		}
		a.cleaner = cleaner.New(newDetectors)
		// Reload
		a.state = stateLoading
		return a, a.startLoading()

	case settingsCanceledMsg:
		a.state = stateList
		return a, nil

	case progressTickMsg:
		if a.state == stateLoading {
			var cmd tea.Cmd
			a.progress, cmd = a.progress.Update(msg)
			return a, cmd
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}

	// Delegate to sub-models
	switch a.state {
	case stateList:
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch {
			case key.Matches(keyMsg, a.keys.Confirm):
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
			case key.Matches(keyMsg, a.keys.Detail):
				if item, ok := a.list.list.SelectedItem().(pkg.Package); ok {
					a.currentPkg = item
					a.detail = newDetailModel(item, a.styles, a.width, a.height)
					a.state = stateDetail
					return a, nil
				}
			case key.Matches(keyMsg, a.keys.Settings):
				activeNames := make(map[string]bool, len(a.config.Detectors))
				for _, d := range a.config.Detectors {
					activeNames[d.Name()] = true
				}
				a.settings = newSettingsModel(a.allDetectors, activeNames, a.styles, a.width, a.height)
				a.state = stateSettings
				return a, a.settings.Init()
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
		var cmd tea.Cmd
		a.confirm, cmd = a.confirm.Update(msg)
		return a, cmd

	case stateResult:
		// Let viewport scroll in result view
		if _, ok := msg.(tea.KeyPressMsg); ok {
			var cmd tea.Cmd
			a.resultViewport, cmd = a.resultViewport.Update(msg)
			return a, cmd
		}

	case stateSettings:
		var cmd tea.Cmd
		a.settings, cmd = a.settings.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) View() tea.View {
	var content string
	switch a.state {
	case stateLoading:
		content = a.splashView()
	case stateList:
		content = a.list.View()
	case stateDetail:
		content = a.detail.View()
	case stateConfirm:
		content = a.confirm.View()
	case stateUninstalling:
		if a.sudoCurrentPkg != nil {
			content = fmt.Sprintf("\n  Uninstalling %s (requires sudo)...\n\n  The terminal will prompt for your password.\n",
				a.sudoCurrentPkg.Name)
		} else {
			content = fmt.Sprintf("\n  %s Uninstalling packages...\n", a.spinner.View())
		}
	case stateResult:
		content = a.resultView()
	case stateSettings:
		content = a.settings.View()
	case stateError:
		content = a.styles.ErrorText.Render(fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.", a.err))
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// --- Commands ---

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
		a.buildResultTable()
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

// buildResultTable constructs a table model from uninstall results.
func (a *App) buildResultTable() {
	columns := []table.Column{
		{Title: "Status", Width: 6},
		{Title: "Manager", Width: 10},
		{Title: "Package", Width: 30},
		{Title: "Details", Width: 40},
	}

	var rows []table.Row
	for _, r := range a.results {
		var status, detail string
		switch {
		case r.DryRunCmd != "":
			status = "DRY"
			detail = r.DryRunCmd
		case r.Err != nil:
			status = "✗"
			detail = r.Err.Error()
		default:
			status = "✓"
			if len(r.Leftovers) > 0 {
				detail = "leftovers: " + strings.Join(r.Leftovers, ", ")
			} else {
				detail = "clean"
			}
		}
		rows = append(rows, table.Row{status, string(r.Package.Manager), r.Package.Name, detail})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)+1, a.height-10)),
	)
	t.SetStyles(a.styles.TableStyles)
	a.resultTable = t
}

func (a *App) resultView() string {
	var b strings.Builder

	fmt.Fprintln(&b, a.styles.Title.Render(" 📋 Uninstall Results "))
	fmt.Fprintln(&b)

	// Summary counts
	successCount, failedCount, dryRunCount := 0, 0, 0
	for _, r := range a.results {
		switch {
		case r.DryRunCmd != "":
			dryRunCount++
		case r.Err != nil:
			failedCount++
		default:
			successCount++
		}
	}

	// Render each result with icons
	for _, r := range a.results {
		badge := a.styles.BadgeFor(string(r.Package.Manager))
		switch {
		case r.DryRunCmd != "":
			dryIcon := a.styles.DryRunBadge.Render("DRY")
			fmt.Fprintf(&b, "  %s %s %s\n", dryIcon, badge, r.Package.Name)
			fmt.Fprintf(&b, "    %s\n", a.styles.HelpBar.Render(r.DryRunCmd))

		case r.Err != nil:
			fmt.Fprintf(&b, "  %s %s %s\n",
				a.styles.ErrorText.Render("✗"), badge, r.Package.Name)
			fmt.Fprintf(&b, "    %s\n", a.styles.ErrorText.Render(r.Err.Error()))

		default:
			okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess))
			fmt.Fprintf(&b, "  %s %s %s\n", okStyle.Render("✓"), badge, r.Package.Name)
			if len(r.Leftovers) > 0 {
				fmt.Fprintf(&b, "    %s\n",
					a.styles.HelpBar.Render("Leftover files: "+strings.Join(r.Leftovers, ", ")))
			}
		}
	}

	fmt.Fprintln(&b)

	// Summary bar
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
	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 1).
		Bold(true)
	fmt.Fprintln(&b, summaryStyle.Render(summary))

	// Help
	resultKeys := DefaultResultKeyMap()
	fmt.Fprint(&b, "\n  "+a.help.View(resultKeys))

	return b.String()
}

// splashView renders the branded loading screen as two stacked panels:
// a brand card on top and a detection progress card below. Layout
// adapts to terminal width but keeps a comfortable max so the splash
// does not stretch awkwardly on wide terminals.
func (a *App) splashView() string {
	const maxWidth = 64
	width := a.width - 4
	if width > maxWidth {
		width = maxWidth
	}
	if width < 32 {
		width = 32
	}

	brandTitle := a.theme.Subtitle.Render("✦  OmniClean  ✦")
	tagline := a.theme.Subtle.Render("Clean up your system, one package at a time")
	brandBody := lipgloss.JoinVertical(lipgloss.Center,
		brandTitle, "", tagline,
	)
	brandCard := components.Panel(a.theme, "", lipgloss.NewStyle().
		Width(width-4).
		Align(lipgloss.Center).
		Render(brandBody),
		components.PanelOpts{Width: width, Accent: true},
	)

	// Detection list
	var rows []string
	for i, name := range a.detectorNames {
		switch {
		case i < a.doneCount:
			rows = append(rows, "  "+a.theme.Success.Render("✓")+"  "+a.theme.Dim.Render(name))
		case i == a.doneCount:
			rows = append(rows, "  "+a.spinner.View()+"  "+a.theme.Subtitle.Render(name)+a.theme.Dim.Render(" …"))
		default:
			rows = append(rows, "  "+a.theme.Dim.Render("○  "+name))
		}
	}
	detectionList := strings.Join(rows, "\n")
	progressBar := a.progress.View()
	countLine := a.theme.Dim.Render(fmt.Sprintf("%d packages found so far", len(a.loadPackages)))
	if len(a.loadPackages) == 0 {
		countLine = a.theme.Dim.Render("scanning…")
	}
	detectionBody := lipgloss.JoinVertical(lipgloss.Left,
		detectionList,
		"",
		progressBar,
		countLine,
	)
	detectionCard := components.Panel(a.theme, " Detecting package managers ", detectionBody,
		components.PanelOpts{Width: width, Accent: false},
	)

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "ctrl+c", Action: "quit"},
	})

	content := lipgloss.JoinVertical(lipgloss.Center,
		brandCard,
		"",
		detectionCard,
		"",
		footer,
	)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

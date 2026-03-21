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
}

// New creates the App model ready to be run.
func New(cfg Config) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	return &App{
		config:  cfg,
		state:   stateLoading,
		styles:  DefaultStyles(),
		spinner: s,
		cleaner: cleaner.New(cfg.Detectors),
	}
}

// Run starts the Bubbletea program.
func (a *App) Run(ctx context.Context) error {
	p := tea.NewProgram(a, tea.WithAltScreen(), tea.WithContext(ctx))
	_, err := p.Run()
	return err
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
		a.list = newListModel(a.packages, a.styles)
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
		return a, a.uninstallCmd(selected)

	case confirmNoMsg:
		a.state = stateList
		return a, nil

	case UninstallCompleteMsg:
		a.results = msg.Results
		a.state = stateResult
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
						return a, a.uninstallCmd(selected)
					}
					a.confirm = newConfirmModel(selected, a.config.DryRun, a.styles)
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
		for _, d := range detectors {
			pkgs, err := d.ListPackages(ctx)
			if err != nil {
				// Log the error but continue with other detectors
				continue
			}
			packages = append(packages, pkgs...)
		}
		return PackagesLoadedMsg{Packages: packages}
	}
}

func (a *App) uninstallCmd(packages []pkg.Package) tea.Cmd {
	cl := a.cleaner
	dryRun := a.config.DryRun
	return func() tea.Msg {
		ctx := context.Background()
		results := cl.Uninstall(ctx, packages, dryRun)
		return UninstallCompleteMsg{Results: results}
	}
}

// --- Views ---

func (a *App) resultView() string {
	var b strings.Builder
	fmt.Fprintln(&b, a.styles.Title.Render("Uninstall Results"))
	fmt.Fprintln(&b)

	success, failed := 0, 0
	for _, r := range a.results {
		badge := a.styles.BadgeFor(string(r.Package.Manager))
		if r.Err != nil {
			fmt.Fprintf(&b, "  %s %s %s\n",
				a.styles.ErrorText.Render("✗"),
				badge, r.Package.Name,
			)
			fmt.Fprintf(&b, "    %s\n", a.styles.ErrorText.Render(r.Err.Error()))
			failed++
		} else {
			ok := lipgloss.NewStyle().Foreground(lipgloss.Color("#48BB78")).Render("✓")
			fmt.Fprintf(&b, "  %s %s %s\n", ok, badge, r.Package.Name)
			success++
		}
	}

	fmt.Fprintln(&b)
	summary := fmt.Sprintf("Done: %d removed", success)
	if failed > 0 {
		summary += fmt.Sprintf(", %d failed", failed)
	}
	fmt.Fprintln(&b, a.styles.StatusBar.Render(summary))
	fmt.Fprint(&b, a.styles.HelpBar.Render("\n  Press q to quit"))

	return b.String()
}

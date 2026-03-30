package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/detector"
)

// settingsModel wraps a huh.Form for configuring active detectors.
type settingsModel struct {
	form      *huh.Form
	selected  []string           // bound to huh MultiSelect
	detectors []detector.Detector // all available detectors
	styles    Styles
	width     int
	height    int
}

func newSettingsModel(detectors []detector.Detector, activeNames map[string]bool, styles Styles, width, height int) settingsModel {
	// Build options — pre-select currently active detectors
	options := make([]huh.Option[string], 0, len(detectors))
	var preSelected []string
	for _, d := range detectors {
		opt := huh.NewOption(formatDetectorLabel(d), d.Name())
		if activeNames[d.Name()] {
			opt = opt.Selected(true)
			preSelected = append(preSelected, d.Name())
		}
		options = append(options, opt)
	}

	selected := make([]string, len(preSelected))
	copy(selected, preSelected)

	m := settingsModel{
		selected:  selected,
		detectors: detectors,
		styles:    styles,
		width:     width,
		height:    height,
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Package Managers").
				Description("Toggle which managers to scan. Press Space to select, Enter to apply.").
				Options(options...).
				Value(&m.selected),
		),
	).WithTheme(huh.ThemeFunc(huh.ThemeCharm)).
		WithWidth(min(width-4, 70)).
		WithHeight(min(height-6, 20))

	m.form = form
	return m
}

func (m settingsModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m settingsModel) Update(msg tea.Msg) (settingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to huh form
	_, cmd := m.form.Update(msg)

	// Check if form completed or aborted
	switch m.form.State {
	case huh.StateCompleted:
		selected := make([]string, len(m.selected))
		copy(selected, m.selected)
		return m, func() tea.Msg {
			return settingsAppliedMsg{SelectedManagers: selected}
		}
	case huh.StateAborted:
		return m, func() tea.Msg {
			return settingsCanceledMsg{}
		}
	}

	return m, cmd
}

func (m settingsModel) View() string {
	var b strings.Builder

	// Header
	title := m.styles.Title.Render(" ⚙  Settings ")
	fmt.Fprintln(&b, title)
	fmt.Fprintln(&b)

	// Form
	fmt.Fprintln(&b, m.form.View())

	// Help
	helpText := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim)).Render(
		"  space: toggle  ·  enter: apply  ·  esc: cancel",
	)
	fmt.Fprintln(&b, helpText)

	return b.String()
}

// formatDetectorLabel creates a display label for a detector.
func formatDetectorLabel(d detector.Detector) string {
	desc := ""
	switch d.Name() {
	case "apt":
		desc = "Debian/Ubuntu system packages"
	case "snap":
		desc = "Snap containerized apps"
	case "flatpak":
		desc = "Flatpak sandboxed apps"
	case "brew":
		desc = "Homebrew formulae & casks"
	case "pip":
		desc = "Python packages"
	case "npm":
		desc = "Node.js global packages"
	case "cargo":
		desc = "Rust crates"
	case "winget":
		desc = "Windows Package Manager"
	case "choco":
		desc = "Chocolatey packages"
	case "scoop":
		desc = "Scoop bucket apps"
	default:
		desc = d.Name()
	}
	return fmt.Sprintf("%s — %s", d.Name(), desc)
}

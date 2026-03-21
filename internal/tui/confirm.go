package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// confirmModel shows a confirmation dialog listing packages to be removed.
type confirmModel struct {
	packages []pkg.Package
	dryRun   bool
	styles   Styles
	width    int
}

func newConfirmModel(packages []pkg.Package, dryRun bool, styles Styles) confirmModel {
	return confirmModel{packages: packages, dryRun: dryRun, styles: styles}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	if sz, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = sz.Width
	}
	return m, nil
}

// confirmYesMsg is returned when user presses y.
type confirmYesMsg struct{}

// confirmNoMsg is returned when user presses n or esc.
type confirmNoMsg struct{}

func (m confirmModel) HandleKey(key string) tea.Cmd {
	switch strings.ToLower(key) {
	case "y", "enter":
		return func() tea.Msg { return confirmYesMsg{} }
	case "n", "esc", "q":
		return func() tea.Msg { return confirmNoMsg{} }
	}
	return nil
}

func (m confirmModel) View() string {
	var b strings.Builder

	header := "Packages to remove:"
	if m.dryRun {
		header += "  " + m.styles.DryRunBadge.Render("DRY RUN")
	}
	fmt.Fprintln(&b, m.styles.Title.Render(header))
	fmt.Fprintln(&b)

	for _, p := range m.packages {
		badge := m.styles.BadgeFor(string(p.Manager))
		fmt.Fprintf(&b, "  %s %s %s\n", badge, p.Name, p.Version)
	}

	fmt.Fprintln(&b)
	if m.dryRun {
		fmt.Fprintln(&b, m.styles.DryRunBadge.Render(" DRY RUN — no changes will be made "))
		fmt.Fprintln(&b)
	}
	fmt.Fprint(&b, m.styles.HelpBar.Render("  [y] confirm  [n/esc] cancel"))

	return m.styles.Border.Width(m.width - 4).Render(b.String())
}

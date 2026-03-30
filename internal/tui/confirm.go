package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// confirmModel shows a confirmation dialog listing packages to be removed.
// Uses styled rendering with the OmniClean design system.
type confirmModel struct {
	packages []pkg.Package
	dryRun   bool
	hasSudo  bool
	styles   Styles
	width    int
}

func newConfirmModel(packages []pkg.Package, dryRun, hasSudo bool, styles Styles, width int) confirmModel {
	return confirmModel{
		packages: packages,
		dryRun:   dryRun,
		hasSudo:  hasSudo,
		styles:   styles,
		width:    width,
	}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (confirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyPressMsg:
		cmd := m.HandleKey(msg.String())
		if cmd != nil {
			return m, cmd
		}
	}
	return m, nil
}

// confirmYesMsg is returned when user presses y.
type confirmYesMsg struct{}

// confirmNoMsg is returned when user presses n or esc.
type confirmNoMsg struct{}

func (m confirmModel) HandleKey(keyStr string) tea.Cmd {
	switch strings.ToLower(keyStr) {
	case "y", "enter":
		return func() tea.Msg { return confirmYesMsg{} }
	case "n", "esc", "q":
		return func() tea.Msg { return confirmNoMsg{} }
	}
	return nil
}

func (m confirmModel) View() string {
	var b strings.Builder

	// Title bar
	titleIcon := "🗑"
	if m.dryRun {
		titleIcon = "👁"
	}
	header := fmt.Sprintf(" %s Packages to remove (%d) ", titleIcon, len(m.packages))
	if m.dryRun {
		header += " " + m.styles.DryRunBadge.Render("DRY RUN")
	}
	fmt.Fprintln(&b, m.styles.Title.Render(header))
	fmt.Fprintln(&b)

	// Package list with badges
	for _, p := range m.packages {
		// Bullet
		bulletCol := "  • "

		// Badge
		badgeCol := lipgloss.NewStyle().Width(ColWidthBadge).Render(m.styles.BadgeFor(string(p.Manager)))

		// Name (Flexible)
		fixedWidths := lipgloss.Width(bulletCol) + ColWidthBadge + ColWidthVersion + ColWidthSize + 4 // spacing
		nameWidth := m.width - 4 /* box padding */ - fixedWidths
		if nameWidth < 10 {
			nameWidth = 10
		}
		nameText := p.Name
		if len(nameText) > nameWidth {
			nameText = truncate.StringWithTail(nameText, uint(nameWidth), "…")
		}
		nameCol := lipgloss.NewStyle().Width(nameWidth).Bold(true).Render(nameText)

		// Version
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim))
		versionText := p.Version
		if len(versionText) > ColWidthVersion {
			versionText = truncate.StringWithTail(versionText, uint(ColWidthVersion), "…")
		}
		versionCol := lipgloss.NewStyle().Width(ColWidthVersion).Render(dimStyle.Render(versionText))

		// Size
		sizeStr := ""
		if p.Size > 0 {
			sizeStr = formatBytes(p.Size)
		}
		sizeCol := lipgloss.NewStyle().Width(ColWidthSize).Align(lipgloss.Right).Render(dimStyle.Render(sizeStr))

		// Render full string safely joined without manual string interpolation artifacts
		row := lipgloss.JoinHorizontal(lipgloss.Left,
			bulletCol,
			badgeCol,
			nameCol,
			"  ",
			versionCol,
			"  ",
			sizeCol,
		)
		fmt.Fprintln(&b, row)
	}

	fmt.Fprintln(&b)

	// Warnings
	if m.dryRun {
		warnBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color(ColorWarning)).
			Padding(0, 1).
			Bold(true).
			Render(" 👁  DRY RUN — no changes will be made ")
		fmt.Fprintln(&b, "  "+warnBox)
		fmt.Fprintln(&b)
	} else if m.hasSudo {
		warnBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color(ColorWarning)).
			Padding(0, 1).
			Bold(true).
			Render(" ⚠  Some packages require sudo access ")
		fmt.Fprintln(&b, "  "+warnBox)
		fmt.Fprintln(&b)
	}

	// Action buttons
	yesStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(ColorSuccess)).
		Padding(0, 2).
		Bold(true)
	noStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(ColorError)).
		Padding(0, 2).
		Bold(true)

	buttons := fmt.Sprintf("  %s  %s",
		yesStyle.Render("y  confirm"),
		noStyle.Render("n  cancel"),
	)
	fmt.Fprintln(&b, buttons)

	width := m.width - 4
	if width < 10 {
		width = 10
	}
	return m.styles.Border.Width(width).Render(b.String())
}

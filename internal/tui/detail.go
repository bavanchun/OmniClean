package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// detailModel shows full information for a single package.
// Uses glamour for markdown rendering and viewport for scrolling.
type detailModel struct {
	pkg      pkg.Package
	styles   Styles
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

func newDetailModel(p pkg.Package, styles Styles, width, height int) detailModel {
	vp := viewport.New(
		viewport.WithWidth(max(width-6, 20)),
		viewport.WithHeight(max(height-8, 10)),
	)
	vp.SetContent(renderDetailContent(p, styles))

	return detailModel{
		pkg:      p,
		styles:   styles,
		viewport: vp,
		width:    width,
		height:   height,
		ready:    true,
	}
}

func (m detailModel) Init() tea.Cmd { return nil }

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(max(msg.Width-6, 20))
		m.viewport.SetHeight(max(msg.Height-8, 10))
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m detailModel) View() string {
	title := m.styles.Title.Render(fmt.Sprintf(" 📦 %s ", m.pkg.Name))
	badge := m.styles.BadgeFor(string(m.pkg.Manager))
	header := fmt.Sprintf("%s  %s", title, badge)

	content := m.viewport.View()

	scrollInfo := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim)).Render(
		fmt.Sprintf("  ↑↓ scroll  ·  %d%%", int(m.viewport.ScrollPercent()*100)),
	)

	help := m.styles.HelpBar.Render("  esc: back  ·  space: select for removal")

	return fmt.Sprintf("%s\n\n%s\n%s\n%s", header, content, scrollInfo, help)
}

// renderDetailContent builds the detail view content, using glamour for
// markdown descriptions when available.
func renderDetailContent(p pkg.Package, styles Styles) string {
	var b strings.Builder

	// Basic info section
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Bold(true)
	valueStyle := lipgloss.NewStyle().Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim))

	fmt.Fprintf(&b, "  %s  %s\n", labelStyle.Render("Version:"), valueStyle.Render(p.Version))
	fmt.Fprintf(&b, "  %s  %s\n", labelStyle.Render("Manager:"), valueStyle.Render(string(p.Manager)))

	if p.Size > 0 {
		fmt.Fprintf(&b, "  %s     %s\n", labelStyle.Render("Size:"), valueStyle.Render(formatBytes(p.Size)))
	}

	if p.Description != "" {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "  %s\n", labelStyle.Render("Description:"))
		fmt.Fprintln(&b)

		// Try to render description as markdown via glamour
		rendered, err := glamour.Render(p.Description, "dark")
		if err == nil && strings.TrimSpace(rendered) != "" {
			fmt.Fprint(&b, rendered)
		} else {
			// Fallback to plain text
			fmt.Fprintf(&b, "  %s\n", dimStyle.Render(p.Description))
		}
	}

	return b.String()
}

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

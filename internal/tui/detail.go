package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// detailModel shows full information for a single package.
type detailModel struct {
	pkg    pkg.Package
	styles Styles
	width  int
	height int
}

func newDetailModel(p pkg.Package, styles Styles) detailModel {
	return detailModel{pkg: p, styles: styles}
}

func (m detailModel) Init() tea.Cmd { return nil }

func (m detailModel) Update(msg tea.Msg) (detailModel, tea.Cmd) {
	if sz, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = sz.Width
		m.height = sz.Height
	}
	return m, nil
}

func (m detailModel) View() string {
	p := m.pkg
	badge := m.styles.BadgeFor(string(p.Manager))

	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s\n\n", m.styles.Title.Render(p.Name), badge)
	fmt.Fprintf(&b, "  Version:  %s\n", lipgloss.NewStyle().Bold(true).Render(p.Version))
	if p.Description != "" {
		fmt.Fprintf(&b, "  Summary:  %s\n", p.Description)
	}
	if p.Size > 0 {
		fmt.Fprintf(&b, "  Size:     %s\n", formatBytes(p.Size))
	}
	fmt.Fprintln(&b)
	fmt.Fprint(&b, m.styles.HelpBar.Render("  esc: back  enter: select for removal"))

	return m.styles.Border.Width(m.width - 4).Render(b.String())
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

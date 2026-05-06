package components

import (
	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// PanelOpts tunes how Panel composes the surrounding box. Width <= 0 lets
// Lipgloss size to content. Accent toggles the brand-colored border.
type PanelOpts struct {
	Width  int
	Accent bool
}

// Panel renders a titled rectangular surface with optional brand-accented
// border. The title strip sits flush against the top edge so panels stack
// cleanly when joined vertically.
func Panel(s theme.Styles, title, body string, opts PanelOpts) string {
	border := s.Border
	if opts.Accent {
		border = s.PanelBorder
	}
	if opts.Width > 0 {
		border = border.Width(opts.Width)
	}

	if title == "" {
		return border.Render(body)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextStrong)).
		Background(lipgloss.Color(theme.Primary)).
		Padding(0, theme.Space1).
		Bold(true)

	header := titleStyle.Render(title)
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	return border.Render(content)
}

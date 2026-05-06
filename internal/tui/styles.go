// Package tui implements the Bubbletea TUI for OmniClean.
package tui

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// Color palette aliases. The canonical palette lives in the theme package.
// These aliases are kept so existing call sites within `tui` keep compiling
// while we migrate views to reference `theme.*` directly.
const (
	ColorPrimary    = theme.Primary
	ColorSuccess    = theme.Success
	ColorWarning    = theme.Warning
	ColorError      = theme.Error
	ColorDim        = theme.TextDim
	ColorSubtle     = theme.TextSubtle
	ColorAccent     = theme.PrimarySoft
	ColorBg         = theme.SurfaceBg
	ColorBorder     = theme.SurfaceBorder
	ColorProgressA  = theme.ProgressStart
	ColorProgressB  = theme.ProgressEnd
	ColorSelectedBg = theme.SurfaceElevated
)

// Fixed column widths for List and Confirm views.
const (
	ColWidthCheckbox = theme.ColCheckbox
	ColWidthBadge    = theme.ColBadge
	ColWidthVersion  = theme.ColVersion
	ColWidthSize     = theme.ColSize
)

// Styles holds all Lipgloss styles used in the application.
type Styles struct {
	Title        lipgloss.Style
	StatusBar    lipgloss.Style
	HelpBar      lipgloss.Style
	ManagerBadge map[string]lipgloss.Style
	SelectedText lipgloss.Style
	SelectedRow  lipgloss.Style
	DryRunBadge  lipgloss.Style
	ErrorText    lipgloss.Style
	Border       lipgloss.Style
	// Progress bar styles
	ProgressBar lipgloss.Style
	// Table styles
	TableStyles table.Styles
	// Viewport border
	ViewportBorder lipgloss.Style
}

// DefaultStyles returns the application's default style set.
func DefaultStyles() Styles {
	managerBadges := make(map[string]lipgloss.Style, len(theme.ManagerColor))
	for name, hex := range theme.ManagerColor {
		managerBadges[name] = lipgloss.NewStyle().
			Foreground(lipgloss.Color(hex)).
			Bold(true)
	}

	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.TextStrong)).
			Background(lipgloss.Color(theme.Primary)).
			Padding(0, theme.Space1),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextSubtle)).
			Padding(0, theme.Space1),

		HelpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextDim)),

		SelectedText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true),

		SelectedRow: lipgloss.NewStyle().
			Background(lipgloss.Color(theme.SurfaceElevated)),

		DryRunBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Black)).
			Background(lipgloss.Color(theme.Warning)).
			Padding(0, theme.Space1).
			Bold(true),

		ErrorText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Error)),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.SurfaceBorder)),

		ProgressBar: lipgloss.NewStyle().
			Padding(theme.Space1, theme.Space2),

		ViewportBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Primary)).
			Padding(0, theme.Space1),

		TableStyles: defaultTableStyles(),

		ManagerBadge: managerBadges,
	}
}

// BadgeFor returns a styled manager name. Falls back to plain text if the
// manager has no defined color.
func (s Styles) BadgeFor(manager string) string {
	if style, ok := s.ManagerBadge[manager]; ok {
		return style.Render("[" + manager + "]")
	}
	return "[" + manager + "]"
}

// defaultTableStyles returns styled table configuration for result views.
func defaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(theme.SurfaceBorder)).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color(theme.Primary))

	s.Selected = s.Selected.
		Foreground(lipgloss.Color(theme.TextStrong)).
		Background(lipgloss.Color(theme.Primary)).
		Bold(true)

	return s
}

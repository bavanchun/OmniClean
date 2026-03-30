// Package tui implements the Bubbletea TUI for OmniClean.
package tui

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

// Color palette constants for OmniClean.
const (
	ColorPrimary    = "#7C3AED" // violet
	ColorSuccess    = "#48BB78" // green
	ColorWarning    = "#F6AD55" // orange
	ColorError      = "#FC8181" // red
	ColorDim        = "#718096" // gray
	ColorSubtle     = "#A0AEC0" // light gray
	ColorAccent     = "#E9D8FD" // lavender
	ColorBg         = "#1A202C" // dark bg
	ColorBorder     = "#4A5568" // border gray
	ColorProgressA  = "#7C3AED" // progress gradient start
	ColorProgressB  = "#E9D8FD" // progress gradient end
)

// Styles holds all Lipgloss styles used in the application.
type Styles struct {
	Title        lipgloss.Style
	StatusBar    lipgloss.Style
	HelpBar      lipgloss.Style
	ManagerBadge map[string]lipgloss.Style
	Selected     lipgloss.Style
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
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(ColorPrimary)).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSubtle)).
			Padding(0, 1),

		HelpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDim)),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true),

		DryRunBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color(ColorWarning)).
			Padding(0, 1).
			Bold(true),

		ErrorText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)),

		ProgressBar: lipgloss.NewStyle().
			Padding(1, 2),

		ViewportBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorPrimary)).
			Padding(0, 1),

		TableStyles: defaultTableStyles(),

		ManagerBadge: map[string]lipgloss.Style{
			// Linux
			"apt":     lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Bold(true),
			"snap":    lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Bold(true),
			"flatpak": lipgloss.NewStyle().Foreground(lipgloss.Color("#63B3ED")).Bold(true),
			// macOS + Linux
			"brew": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Bold(true),
			// Cross-platform
			"pip":   lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")).Bold(true),
			"npm":   lipgloss.NewStyle().Foreground(lipgloss.Color("#F687B3")).Bold(true),
			"cargo": lipgloss.NewStyle().Foreground(lipgloss.Color("#D69E2E")).Bold(true),
			// Windows
			"winget": lipgloss.NewStyle().Foreground(lipgloss.Color("#76E4F7")).Bold(true),
			"choco":  lipgloss.NewStyle().Foreground(lipgloss.Color("#FBD38D")).Bold(true),
			"scoop":  lipgloss.NewStyle().Foreground(lipgloss.Color("#B794F4")).Bold(true),
		},
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
		BorderForeground(lipgloss.Color(ColorBorder)).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimary))

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(ColorPrimary)).
		Bold(true)

	return s
}

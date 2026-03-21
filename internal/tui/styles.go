// Package tui implements the Bubbletea TUI for OmniClean.
package tui

import "github.com/charmbracelet/lipgloss"

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
}

// DefaultStyles returns the application's default style set.
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0AEC0")).
			Padding(0, 1),

		HelpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#718096")),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true),

		DryRunBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#F6AD55")).
			Padding(0, 1).
			Bold(true),

		ErrorText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FC8181")),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A5568")),

		ManagerBadge: map[string]lipgloss.Style{
			// Linux
			"apt":     lipgloss.NewStyle().Foreground(lipgloss.Color("#48BB78")).Bold(true),
			"snap":    lipgloss.NewStyle().Foreground(lipgloss.Color("#FC8181")).Bold(true),
			"flatpak": lipgloss.NewStyle().Foreground(lipgloss.Color("#63B3ED")).Bold(true),
			// macOS + Linux
			"brew": lipgloss.NewStyle().Foreground(lipgloss.Color("#F6AD55")).Bold(true),
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

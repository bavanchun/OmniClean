package theme

import "charm.land/lipgloss/v2"

// Styles is the canonical preset of Lipgloss styles built on top of the
// design tokens. Views must consume this set instead of constructing
// styles inline so look-and-feel stays consistent.
type Styles struct {
	// Typography
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Body     lipgloss.Style
	Dim      lipgloss.Style
	Subtle   lipgloss.Style
	Strong   lipgloss.Style

	// Status text
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Info    lipgloss.Style

	// Bars
	TitleBar  lipgloss.Style
	StatusBar lipgloss.Style
	HelpBar   lipgloss.Style
	HelpKey   lipgloss.Style

	// Selection
	SelectedRow  lipgloss.Style
	SelectedText lipgloss.Style

	// Surfaces
	Border         lipgloss.Style
	PanelBorder    lipgloss.Style
	ViewportBorder lipgloss.Style

	// Badges
	BadgeDryRun  lipgloss.Style
	BadgeSuccess lipgloss.Style
	BadgeWarning lipgloss.Style
	BadgeError   lipgloss.Style
	BadgeInfo    lipgloss.Style
	ManagerBadge map[string]lipgloss.Style
}

// New builds the preset using the current token palette. Callers can keep
// a single instance for the lifetime of a Bubbletea program; styles are
// immutable Lipgloss values so concurrent reads are safe.
func New() Styles {
	managerBadges := make(map[string]lipgloss.Style, len(ManagerColor))
	for name, hex := range ManagerColor {
		managerBadges[name] = lipgloss.NewStyle().
			Foreground(lipgloss.Color(hex)).
			Bold(true)
	}

	pill := func(bg string, fg string) lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(fg)).
			Background(lipgloss.Color(bg)).
			Padding(0, Space1).
			Bold(true)
	}

	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(TextStrong)).
			Background(lipgloss.Color(Primary)).
			Padding(0, Space1),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimarySoft)).
			Bold(true),

		Body: lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextBody)),

		Dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextDim)),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextSubtle)),

		Strong: lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextStrong)).
			Bold(true),

		Success: lipgloss.NewStyle().Foreground(lipgloss.Color(Success)).Bold(true),
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color(Warning)).Bold(true),
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color(Error)).Bold(true),
		Info:    lipgloss.NewStyle().Foreground(lipgloss.Color(Info)).Bold(true),

		TitleBar: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(TextStrong)).
			Background(lipgloss.Color(Primary)).
			Padding(0, Space2),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextSubtle)).
			Padding(0, Space1),

		HelpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextDim)),

		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimarySoft)).
			Bold(true),

		SelectedRow: lipgloss.NewStyle().
			Background(lipgloss.Color(SurfaceElevated)),

		SelectedText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(Primary)).
			Bold(true),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(SurfaceBorder)),

		PanelBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Primary)).
			Padding(Space1, Space2),

		ViewportBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(Primary)).
			Padding(0, Space1),

		BadgeDryRun:  pill(Warning, "#000000"),
		BadgeSuccess: pill(Success, TextStrong),
		BadgeWarning: pill(Warning, "#000000"),
		BadgeError:   pill(Error, TextStrong),
		BadgeInfo:    pill(Info, TextStrong),
		ManagerBadge: managerBadges,
	}
}

// Manager returns the styled manager badge label, falling back to plain
// brackets when a manager is unknown so render code remains panic-free.
func (s Styles) Manager(name string) string {
	if style, ok := s.ManagerBadge[name]; ok {
		return style.Render("[" + name + "]")
	}
	return "[" + name + "]"
}

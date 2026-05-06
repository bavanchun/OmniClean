package menu

import (
	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextSubtle))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Primary)).
			Padding(0, 2)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true)

	activeTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextStrong)).
				Bold(true)

	activeDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextSubtle))

	inactiveTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextSubtle))

	inactiveDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextMuted))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextDim))

	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true)

	brandPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.BrandPanelBg)).
			Padding(1, 2)

	brandTaglineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.TextSubtle))

	brandMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextDim))

	cardBase = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 2).
			Width(46)

	cardActive = cardBase.
			BorderForegroundBlend(
			lipgloss.Color(theme.BorderActive),
			lipgloss.Color(theme.BorderActive2),
		)

	cardIdle = cardBase.
			BorderForeground(lipgloss.Color(theme.BorderIdle))

	barActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.BarActive)).
			Bold(true)

	barIdleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextMuted))

	indexActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true)

	indexIdleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.TextDim))
)

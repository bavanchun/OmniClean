package menu

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"

	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// Adaptive text colors so the menu reads well on both light and dark
// terminals. Borders stay solid so the gradient effect stays predictable.
var (
	textStrong = compat.AdaptiveColor{
		Light: lipgloss.Color("#1A202C"),
		Dark:  lipgloss.Color(theme.TextStrong),
	}
	textBody = compat.AdaptiveColor{
		Light: lipgloss.Color("#2D3748"),
		Dark:  lipgloss.Color(theme.TextBody),
	}
	textSubtle = compat.AdaptiveColor{
		Light: lipgloss.Color("#4A5568"),
		Dark:  lipgloss.Color(theme.TextSubtle),
	}
	textDim = compat.AdaptiveColor{
		Light: lipgloss.Color("#718096"),
		Dark:  lipgloss.Color(theme.TextDim),
	}
	textMuted = compat.AdaptiveColor{
		Light: lipgloss.Color("#A0AEC0"),
		Dark:  lipgloss.Color(theme.TextMuted),
	}
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true)

	activeTitleStyle = lipgloss.NewStyle().
				Foreground(textStrong).
				Bold(true)

	activeDescStyle = lipgloss.NewStyle().
			Foreground(textSubtle)

	inactiveTitleStyle = lipgloss.NewStyle().
				Foreground(textBody)

	inactiveDescStyle = lipgloss.NewStyle().
				Foreground(textMuted)

	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true)

	brandPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.BrandPanelBg)).
			Padding(1, 2)

	brandTaglineStyle = lipgloss.NewStyle().
				Foreground(textSubtle)

	brandMetaStyle = lipgloss.NewStyle().
			Foreground(textDim)

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
			Foreground(textDim)
)

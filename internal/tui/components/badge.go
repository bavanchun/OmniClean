// Package components contains small Lipgloss-based render helpers reused
// across OmniClean TUI views. Components do not own any state; they accept
// the active theme.Styles and return a rendered string.
package components

import (
	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// BadgeKind selects the visual variant used by Badge.
type BadgeKind int

const (
	BadgeNeutral BadgeKind = iota
	BadgeSuccess
	BadgeWarning
	BadgeError
	BadgeInfo
	BadgeDryRun
	BadgeManager
)

// Badge renders a short label as a colored pill or, for BadgeManager,
// delegates to the per-manager styling defined in theme. The label is the
// raw text; brackets and padding are added by the chosen style.
func Badge(s theme.Styles, kind BadgeKind, label string) string {
	switch kind {
	case BadgeSuccess:
		return s.BadgeSuccess.Render(label)
	case BadgeWarning:
		return s.BadgeWarning.Render(label)
	case BadgeError:
		return s.BadgeError.Render(label)
	case BadgeInfo:
		return s.BadgeInfo.Render(label)
	case BadgeDryRun:
		return s.BadgeDryRun.Render(label)
	case BadgeManager:
		return s.Manager(label)
	default:
		return s.Subtle.Render("[" + label + "]")
	}
}

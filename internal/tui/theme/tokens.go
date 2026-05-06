// Package theme centralizes design tokens (colors, spacing, typography,
// radius, layout widths) used across all OmniClean TUI views.
//
// Tokens are deliberately plain string/int constants and small maps so they
// can be consumed by any package without dragging Lipgloss imports.
// Composed Lipgloss styles built on top of these tokens live in styles.go.
package theme

// Brand and semantic palette. Hex strings are kept here as the single
// source of truth; views must reference these names rather than literals.
const (
	Primary   = "#7C3AED"
	PrimarySoft = "#E9D8FD"
	Accent    = "#9F7AEA"

	Success = "#48BB78"
	Warning = "#F6AD55"
	Error   = "#FC8181"
	Info    = "#63B3ED"

	TextStrong = "#FFFFFF"
	TextBody   = "#E2E8F0"
	TextSubtle = "#A0AEC0"
	TextDim    = "#718096"
	TextMuted  = "#4A5568"

	SurfaceBg       = "#1A202C"
	SurfaceElevated = "#2D3748"
	SurfaceBorder   = "#4A5568"

	ProgressStart = Primary
	ProgressEnd   = PrimarySoft
)

// ManagerColor maps each supported package manager to its badge foreground.
// Keep keys in sync with pkg.ManagerType values.
var ManagerColor = map[string]string{
	"apt":     Success,
	"snap":    Error,
	"flatpak": "#63B3ED",
	"brew":    Warning,
	"pip":     "#68D391",
	"npm":     "#F687B3",
	"cargo":   "#D69E2E",
	"winget":  "#76E4F7",
	"choco":   "#FBD38D",
	"scoop":   "#B794F4",
}

// Spacing scale. Values are character cells, designed for Lipgloss
// Padding/Margin where one unit equals one column or one row.
const (
	Space0 = 0
	Space1 = 1
	Space2 = 2
	Space3 = 3
	Space4 = 4
)

// Layout widths used by list/confirm views to keep columns aligned.
const (
	ColCheckbox = 4
	ColBadge    = 11 // longest manager label "[flatpak] " == 10 chars + slack
	ColVersion  = 25
	ColSize     = 10
)

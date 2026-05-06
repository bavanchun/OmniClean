package analyze

import (
	"fmt"
	"strings"
)

// bar renders a horizontal Unicode bar of width characters representing
// pct (0..1). Cells past the value use a dim character so the bar
// width stays constant across rows.
func bar(width int, pct float64) string {
	if width <= 0 {
		return ""
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(float64(width)*pct + 0.5)
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// formatBytes renders a byte count as a compact human string. The
// analyze TUI keeps its own copy so it does not depend on the sibling
// purge package.
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

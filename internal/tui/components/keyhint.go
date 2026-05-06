package components

import (
	"strings"

	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// KeyHint is a single key/action pair shown in the footer help bar.
type KeyHint struct {
	Key    string
	Action string
}

// KeyHints renders a list of KeyHint values into a single dim help line,
// separating entries with a middle dot. Use this for footer help so views
// share the same visual rhythm.
func KeyHints(s theme.Styles, hints []KeyHint) string {
	if len(hints) == 0 {
		return ""
	}
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, s.HelpKey.Render(h.Key)+" "+s.HelpBar.Render(h.Action))
	}
	return s.HelpBar.Render(strings.Join(parts, s.HelpBar.Render("  ·  ")))
}

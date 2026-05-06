package components

import (
	"strings"
	"testing"

	"github.com/bavanchun/OmniClean/internal/tui/theme"
)

// TestPanelRendersTitleAndBody is a render smoke test: it does not
// assert on exact ANSI output (which is fragile across terminals) but
// verifies the rendered string contains the title text and body text
// so future refactors of Panel cannot silently drop them.
func TestPanelRendersTitleAndBody(t *testing.T) {
	s := theme.New()
	out := Panel(s, "Hello", "World inside", PanelOpts{Width: 30, Accent: true})
	if !strings.Contains(out, "Hello") {
		t.Errorf("expected title 'Hello' in output\n%s", out)
	}
	if !strings.Contains(out, "World inside") {
		t.Errorf("expected body 'World inside' in output\n%s", out)
	}
}

func TestBadgeRendersLabel(t *testing.T) {
	s := theme.New()
	cases := []struct {
		kind  BadgeKind
		label string
	}{
		{BadgeSuccess, "OK"},
		{BadgeError, "FAIL"},
		{BadgeWarning, "WARN"},
		{BadgeInfo, "INFO"},
		{BadgeDryRun, "DRY"},
		{BadgeManager, "brew"},
	}
	for _, c := range cases {
		got := Badge(s, c.kind, c.label)
		if !strings.Contains(got, c.label) {
			t.Errorf("expected label %q in badge kind %d output\n%s", c.label, c.kind, got)
		}
	}
}

func TestKeyHintsJoinsEntries(t *testing.T) {
	s := theme.New()
	out := KeyHints(s, []KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "q", Action: "quit"},
	})
	for _, want := range []string{"↑/↓", "navigate", "q", "quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in KeyHints output\n%s", want, out)
		}
	}
}

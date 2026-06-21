package menu

import (
	"runtime"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// renderAt builds a fresh App, feeds a WindowSizeMsg, and returns the
// rendered string. Renders bypass terminal IO; we only assert on
// structural substrings, never raw ANSI.
func renderAt(t *testing.T, w, h, cursor int) string {
	t.Helper()
	a := New(Options{})
	a.cursor = cursor
	model, _ := a.Update(tea.WindowSizeMsg{Width: w, Height: h})
	out := model.(*App)
	return out.render()
}

func TestMenuRender_Wide_CursorTop(t *testing.T) {
	out := renderAt(t, 120, 30, 0)
	for _, want := range []string{
		"Uninstall Packages",
		"Analyze Disk",
		"Purge Project Artifacts",
		"OmniClean",
		"[1]",
		"[2]",
		"[3]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q\n---\n%s", want, out)
		}
	}
	// The active marker should appear at least once.
	if !strings.Contains(out, "▌") {
		t.Errorf("expected active accent bar ▌ in output\n%s", out)
	}
}

func TestMenuRender_Narrow_FallbackSingleColumn(t *testing.T) {
	out := renderAt(t, 60, 40, 0)
	bannerIdx := strings.Index(out, "OmniClean")
	cardIdx := strings.Index(out, "Uninstall Packages")
	if bannerIdx < 0 || cardIdx < 0 {
		t.Fatalf("missing expected substrings\n%s", out)
	}
	if bannerIdx > cardIdx {
		t.Errorf("expected brand panel before cards in single-column fallback")
	}
}

func TestMenuRender_Wide_CursorMid(t *testing.T) {
	out := renderAt(t, 120, 30, 1)
	if !strings.Contains(out, "Analyze Disk") {
		t.Fatalf("missing 'Analyze Disk' in render\n%s", out)
	}
	// The active accent bar must precede the active item's title.
	barIdx := strings.Index(out, "▌")
	analyzeIdx := strings.Index(out, "Analyze Disk")
	if barIdx < 0 || analyzeIdx < 0 || barIdx > analyzeIdx {
		t.Errorf("expected ▌ to appear before Analyze Disk\nbar=%d analyze=%d", barIdx, analyzeIdx)
	}
}

// TestMenuItemSelectionMapping locks the positional enum invariant: the
// Selection value produced for a cursor index is exactly index+1, and the item
// order matches buildMenuItems for the CURRENT platform. This is the guard
// against silent menu mis-routing called out in the plan red-team.
func TestMenuItemSelectionMapping(t *testing.T) {
	items := buildMenuItems()

	// Expected Selection per item index (index+1), independent of platform.
	wantSel := []Selection{SelectUninstall, SelectAnalyze, SelectPurge, SelectCleanup}
	if runtime.GOOS == "darwin" {
		wantSel = append(wantSel, SelectUninstallApps)
	}

	if len(items) != len(wantSel) {
		t.Fatalf("buildMenuItems len = %d, want %d on %s", len(items), len(wantSel), runtime.GOOS)
	}

	// Cursor index i must select wantSel[i] == Selection(i+1).
	for i := range items {
		if got := Selection(i + 1); got != wantSel[i] {
			t.Errorf("index %d (%q) -> Selection %d, want %d", i, items[i].title, got, wantSel[i])
		}
	}

	// "Cleanup Suggestions" must be present on every platform.
	found := false
	for _, it := range items {
		if strings.Contains(it.title, "Cleanup") {
			found = true
		}
	}
	if !found {
		t.Error("Cleanup Suggestions item missing from menu")
	}

	// On darwin, Uninstall Apps must remain the last (highest-ordinal) item.
	if runtime.GOOS == "darwin" {
		last := items[len(items)-1]
		if !strings.Contains(last.title, "Uninstall Apps") {
			t.Errorf("darwin last item = %q, want it to remain 'Uninstall Apps'", last.title)
		}
		if Selection(len(items)) != SelectUninstallApps {
			t.Errorf("SelectUninstallApps ordinal drifted: got %d", Selection(len(items)))
		}
	}
}

func TestMenuJumpKey_SelectsImmediately(t *testing.T) {
	a := New(Options{})
	a.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model, cmd := a.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	out := model.(*App)
	if out.selected != SelectAnalyze {
		t.Errorf("expected SelectAnalyze, got %v", out.selected)
	}
	if cmd == nil {
		t.Errorf("expected tea.Quit cmd after jump")
	}
}

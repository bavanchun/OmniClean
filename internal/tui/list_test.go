package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestSelectionKey(t *testing.T) {
	tests := []struct {
		p    pkg.Package
		want string
	}{
		{pkg.Package{Manager: "apt", Name: "curl"}, "apt:curl"},
		{pkg.Package{Manager: "brew", Name: "git"}, "brew:git"},
		{pkg.Package{Manager: "pip", Name: "requests"}, "pip:requests"},
	}
	for _, tc := range tests {
		got := selectionKey(tc.p)
		if got != tc.want {
			t.Errorf("selectionKey(%v) = %q, want %q", tc.p, got, tc.want)
		}
	}
}

func TestListModel_SelectedPackages_Empty(t *testing.T) {
	m := newListModel(nil, DefaultStyles(), nil)
	selected := m.SelectedPackages()
	if len(selected) != 0 {
		t.Errorf("expected empty selection, got %v", selected)
	}
}

func TestListModel_SelectedPackages(t *testing.T) {
	packages := []pkg.Package{
		{Name: "curl", Manager: "apt", Version: "7.88.1"},
		{Name: "vim", Manager: "apt", Version: "9.0"},
		{Name: "git", Manager: "brew", Version: "2.43.0"},
	}
	m := newListModel(packages, DefaultStyles(), nil)

	// Select curl and git
	m.selected[selectionKey(packages[0])] = true
	m.selected[selectionKey(packages[2])] = true

	selected := m.SelectedPackages()
	if len(selected) != 2 {
		t.Fatalf("got %d selected packages, want 2", len(selected))
	}

	names := map[string]bool{}
	for _, p := range selected {
		names[p.Name] = true
	}
	if !names["curl"] {
		t.Error("expected curl to be selected")
	}
	if !names["git"] {
		t.Error("expected git to be selected")
	}
	if names["vim"] {
		t.Error("vim should not be selected")
	}
}

func TestListModel_SortBySize(t *testing.T) {
	packages := []pkg.Package{
		{Name: "small", Manager: "apt", Size: 1024},
		{Name: "large", Manager: "apt", Size: 1024 * 1024 * 100},
		{Name: "medium", Manager: "apt", Size: 1024 * 1024},
		{Name: "unknown", Manager: "pip", Size: 0},
	}
	m := newListModel(packages, DefaultStyles(), nil)

	// Press 's' to sort by size descending.
	sKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	m, _ = m.Update(sKey)

	if !m.sortedBySize {
		t.Fatal("sortedBySize should be true after pressing s")
	}

	items := m.list.Items()
	if len(items) != 4 {
		t.Fatalf("got %d items, want 4", len(items))
	}

	first, ok := items[0].(pkg.Package)
	if !ok {
		t.Fatal("items[0] is not a Package")
	}
	if first.Name != "large" {
		t.Errorf("items[0] = %q, want %q (largest size first)", first.Name, "large")
	}

	// Press 's' again to restore original order.
	m, _ = m.Update(sKey)
	if m.sortedBySize {
		t.Error("sortedBySize should be false after second press")
	}
	items = m.list.Items()
	first, _ = items[0].(pkg.Package)
	if first.Name != "small" {
		t.Errorf("after restore, items[0] = %q, want %q (original order)", first.Name, "small")
	}
}

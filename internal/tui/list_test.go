package tui

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

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

	// Press 's' to sort by size descending — using key.Binding match.
	sKey := tea.KeyPressMsg(tea.Key{Code: 's'})
	_ = key.Matches(sKey, m.keys.SortSize) // verify this compiles

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

func TestListModel_SelectAll(t *testing.T) {
	packages := []pkg.Package{
		{Name: "curl", Manager: "apt", Version: "7.88.1"},
		{Name: "vim", Manager: "apt", Version: "9.0"},
	}
	m := newListModel(packages, DefaultStyles(), nil)

	// Press 'a' to select all
	aKey := tea.KeyPressMsg(tea.Key{Code: 'a'})
	m, _ = m.Update(aKey)

	selected := m.SelectedPackages()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected after select-all, got %d", len(selected))
	}

	// Press 'a' again to deselect all
	m, _ = m.Update(aKey)
	selected = m.SelectedPackages()
	if len(selected) != 0 {
		t.Errorf("expected 0 selected after second select-all, got %d", len(selected))
	}
}

func TestDefaultKeyMap_HelpInterface(t *testing.T) {
	km := DefaultKeyMap()
	short := km.ShortHelp()
	if len(short) == 0 {
		t.Error("ShortHelp should return non-empty bindings")
	}
	full := km.FullHelp()
	if len(full) == 0 {
		t.Error("FullHelp should return non-empty binding columns")
	}
}

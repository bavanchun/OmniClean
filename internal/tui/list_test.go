package tui

import (
	"testing"

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

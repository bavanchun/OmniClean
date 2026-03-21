package detector

import (
	"testing"
)

func TestAllDetectors_Count(t *testing.T) {
	all := AllDetectors()
	// apt, snap, flatpak, brew, pip, npm, cargo, winget, choco, scoop = 10
	if len(all) != 10 {
		t.Errorf("AllDetectors() returned %d detectors, want 10", len(all))
	}
}

func TestAllDetectors_UniqueNames(t *testing.T) {
	all := AllDetectors()
	seen := make(map[string]bool, len(all))
	for _, d := range all {
		name := d.Name()
		if seen[name] {
			t.Errorf("duplicate detector name %q", name)
		}
		seen[name] = true
	}
}

func TestAllDetectors_ExpectedNames(t *testing.T) {
	want := map[string]bool{
		"apt": true, "snap": true, "flatpak": true, "brew": true,
		"pip": true, "npm": true, "cargo": true,
		"winget": true, "choco": true, "scoop": true,
	}
	for _, d := range AllDetectors() {
		if !want[d.Name()] {
			t.Errorf("unexpected detector name %q", d.Name())
		}
		delete(want, d.Name())
	}
	for name := range want {
		t.Errorf("missing detector %q from AllDetectors()", name)
	}
}

func TestAvailableDetectors_SubsetOfAll(t *testing.T) {
	all := AllDetectors()
	available := AvailableDetectors()

	allNames := make(map[string]bool, len(all))
	for _, d := range all {
		allNames[d.Name()] = true
	}
	for _, d := range available {
		if !allNames[d.Name()] {
			t.Errorf("AvailableDetectors() returned %q which is not in AllDetectors()", d.Name())
		}
	}
	if len(available) > len(all) {
		t.Errorf("AvailableDetectors() (%d) > AllDetectors() (%d)", len(available), len(all))
	}
}

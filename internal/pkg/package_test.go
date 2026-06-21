package pkg

import (
	"testing"
	"time"
)

// TestRole_ZeroValue locks the safety invariant: an unclassified package
// (zero-value Role) must read as RoleUnknown, never as a removable role.
func TestRole_ZeroValue(t *testing.T) {
	var r Role
	if r != RoleUnknown {
		t.Errorf("zero-value Role = %q, want RoleUnknown (%q)", r, RoleUnknown)
	}
	if Role("") != RoleUnknown {
		t.Errorf("Role(\"\") = %q, want RoleUnknown", Role(""))
	}
}

// TestPackage_RoleFieldsAdditive verifies the new Role/InstalledAt fields are
// purely additive: a Package built without them behaves exactly like before.
func TestPackage_RoleFieldsAdditive(t *testing.T) {
	p := Package{Name: "git", Version: "2.43.0", Manager: ManagerBrew}
	if p.Role != RoleUnknown {
		t.Errorf("default Role = %q, want RoleUnknown", p.Role)
	}
	if !p.InstalledAt.IsZero() {
		t.Errorf("default InstalledAt = %v, want zero", p.InstalledAt)
	}
	if p.Desc() != "[brew] 2.43.0" {
		t.Errorf("Desc() = %q, want unchanged %q", p.Desc(), "[brew] 2.43.0")
	}

	// Setting the new fields must not alter Desc rendering.
	p.Role = RoleOrphan
	p.InstalledAt = time.Now()
	if p.Desc() != "[brew] 2.43.0" {
		t.Errorf("Desc() changed after setting Role/InstalledAt: %q", p.Desc())
	}
}

func TestParseHumanSize(t *testing.T) {
	mb := int64(1024 * 1024)
	gb := int64(1024 * 1024 * 1024)
	tb := int64(1024 * 1024 * 1024 * 1024)

	tests := []struct {
		input string
		want  int64
	}{
		{"56.6 MB", int64(566 * mb / 10)},
		{"1.2 GB", int64(12 * gb / 10)},
		{"512 B", 512},
		{"100 KB", 100 * 1024},
		{"100 kB", 100 * 1024},
		{"2.5 TB", int64(25 * tb / 10)},
		{"", 0},
		{"badvalue MB", 0},
		{"100", 0},
		{"-1 MB", 0},
	}
	for _, tc := range tests {
		got := ParseHumanSize(tc.input)
		if got != tc.want {
			t.Errorf("ParseHumanSize(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

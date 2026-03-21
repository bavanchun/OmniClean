package detector

import (
	"context"
	"errors"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestBrew_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name:    "parses two packages",
			output:  "git 2.43.0\nnode 21.6.1\n",
			wantLen: 2,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "git" {
					t.Errorf("got Name %q, want %q", pkgs[0].Name, "git")
				}
				if pkgs[0].Version != "2.43.0" {
					t.Errorf("got Version %q, want %q", pkgs[0].Version, "2.43.0")
				}
				if pkgs[0].Manager != pkg.ManagerBrew {
					t.Errorf("got Manager %q, want %q", pkgs[0].Manager, pkg.ManagerBrew)
				}
			},
		},
		{
			name:    "package with multiple version entries picks last",
			output:  "python@3.11 3.11.7 3.11.6\n",
			wantLen: 1,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Version != "3.11.6" {
					t.Errorf("got Version %q, want last %q", pkgs[0].Version, "3.11.6")
				}
			},
		},
		{
			name:    "empty output returns empty slice",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "command error propagates",
			runErr:  errors.New("brew: command not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewBrew(runner)
			pkgs, err := d.ListPackages(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("ListPackages() error = %v, wantErr %v", err, tc.wantErr)
			}
			if len(pkgs) != tc.wantLen {
				t.Errorf("got %d packages, want %d", len(pkgs), tc.wantLen)
			}
			if tc.check != nil && err == nil {
				tc.check(t, pkgs)
			}
		})
	}
}

func TestBrew_Metadata(t *testing.T) {
	d := NewBrew(DefaultRunner)
	if d.Name() != "brew" {
		t.Errorf("Name() = %q, want %q", d.Name(), "brew")
	}
	if d.NeedsSudo() {
		t.Error("NeedsSudo() = true, want false")
	}
}

package detector

import (
	"context"
	"errors"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestCargo_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name:    "parses two crates",
			output:  "ripgrep v14.1.0:\n    rg\nbat v0.24.0:\n    bat\n",
			wantLen: 2,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "ripgrep" {
					t.Errorf("Name = %q, want %q", pkgs[0].Name, "ripgrep")
				}
				if pkgs[0].Version != "14.1.0" {
					t.Errorf("Version = %q, want %q", pkgs[0].Version, "14.1.0")
				}
				if pkgs[0].Manager != pkg.ManagerCargo {
					t.Errorf("Manager = %q, want %q", pkgs[0].Manager, pkg.ManagerCargo)
				}
			},
		},
		{
			name:    "crate with path suffix",
			output:  "myapp v1.0.0 (/home/user/src/myapp):\n    myapp\n",
			wantLen: 1,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Version != "1.0.0" {
					t.Errorf("Version = %q, want %q", pkgs[0].Version, "1.0.0")
				}
			},
		},
		{
			name:    "empty output",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "command error propagates",
			runErr:  errors.New("cargo: not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewCargo(runner)
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

func TestCargo_Metadata(t *testing.T) {
	d := NewCargo(DefaultRunner)
	if d.Name() != "cargo" {
		t.Errorf("Name() = %q, want %q", d.Name(), "cargo")
	}
	if d.NeedsSudo() {
		t.Error("NeedsSudo() = true, want false")
	}
}

package detector

import (
	"context"
	"errors"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestPip_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name:    "parses freeze format",
			output:  "requests==2.31.0\nnumpy==1.26.3\n",
			wantLen: 2,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "requests" {
					t.Errorf("Name = %q, want %q", pkgs[0].Name, "requests")
				}
				if pkgs[0].Version != "2.31.0" {
					t.Errorf("Version = %q, want %q", pkgs[0].Version, "2.31.0")
				}
				if pkgs[0].Manager != pkg.ManagerPip {
					t.Errorf("Manager = %q, want %q", pkgs[0].Manager, pkg.ManagerPip)
				}
			},
		},
		{
			name:    "skips comment lines",
			output:  "# comment\nrequests==2.31.0\n",
			wantLen: 1,
		},
		{
			name:    "package with no version",
			output:  "somepackage\n",
			wantLen: 1,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Version != "" {
					t.Errorf("Version = %q, want empty", pkgs[0].Version)
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
			runErr:  errors.New("pip: not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewPip(runner)
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

func TestPip_Metadata(t *testing.T) {
	d := NewPip(DefaultRunner)
	if d.Name() != "pip" {
		t.Errorf("Name() = %q, want %q", d.Name(), "pip")
	}
	if d.NeedsSudo() {
		t.Error("NeedsSudo() = true, want false")
	}
}

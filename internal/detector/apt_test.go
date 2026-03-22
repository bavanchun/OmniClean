package detector

import (
	"context"
	"errors"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestAPT_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name: "parses two installed packages",
			output: "ii \tcurl\t7.88.1-10\t712\tcommand line tool for transferring data with URL syntax\n" +
				"ii \tvim\t2:9.0.1378-2\t3716\tVi IMproved - enhanced vi editor\n",
			wantLen: 2,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "curl" {
					t.Errorf("got Name %q, want %q", pkgs[0].Name, "curl")
				}
				if pkgs[0].Version != "7.88.1-10" {
					t.Errorf("got Version %q, want %q", pkgs[0].Version, "7.88.1-10")
				}
				if pkgs[0].Size != 712*1024 {
					t.Errorf("got Size %d, want %d", pkgs[0].Size, 712*1024)
				}
				if pkgs[0].Manager != pkg.ManagerAPT {
					t.Errorf("got Manager %q, want %q", pkgs[0].Manager, pkg.ManagerAPT)
				}
			},
		},
		{
			name: "skips rc (removed, config-files) packages",
			output: "ii \tcurl\t7.88.1-10\t712\tcommand line tool\n" +
				"rc \twps-office\t11.1.0\t1000\tWPS Office\n",
			wantLen: 1,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "curl" {
					t.Errorf("got Name %q, want %q", pkgs[0].Name, "curl")
				}
			},
		},
		{
			name:    "empty output returns empty slice",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "skips blank lines",
			output:  "\n\n",
			wantLen: 0,
		},
		{
			name:    "command error propagates",
			runErr:  errors.New("dpkg-query: not found"),
			wantErr: true,
		},
		{
			name:    "line with only one field is skipped",
			output:  "badline\n",
			wantLen: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewAPT(runner)
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

func TestAPT_Metadata(t *testing.T) {
	d := NewAPT(DefaultRunner)
	if d.Name() != "apt" {
		t.Errorf("Name() = %q, want %q", d.Name(), "apt")
	}
	if !d.NeedsSudo() {
		t.Error("NeedsSudo() = false, want true")
	}
	p := pkg.Package{Name: "curl", Manager: pkg.ManagerAPT}
	cmd := d.DryRunCommand(p)
	if cmd == "" {
		t.Error("DryRunCommand() returned empty string")
	}
}

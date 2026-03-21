package detector

import (
	"context"
	"errors"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestFlatpak_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name:    "parses two apps",
			output:  "org.mozilla.firefox\t121.0\norg.videolan.VLC\t3.0.20\n",
			wantLen: 2,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "org.mozilla.firefox" {
					t.Errorf("Name = %q, want %q", pkgs[0].Name, "org.mozilla.firefox")
				}
				if pkgs[0].Version != "121.0" {
					t.Errorf("Version = %q, want %q", pkgs[0].Version, "121.0")
				}
				if pkgs[0].Manager != pkg.ManagerFlatpak {
					t.Errorf("Manager = %q, want %q", pkgs[0].Manager, pkg.ManagerFlatpak)
				}
			},
		},
		{
			name:    "app with no version",
			output:  "org.example.App\n",
			wantLen: 1,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Version != "" {
					t.Errorf("Version = %q, want empty", pkgs[0].Version)
				}
			},
		},
		{
			name:    "empty output returns empty",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "command error propagates",
			runErr:  errors.New("flatpak: not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewFlatpak(runner)
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

func TestFlatpak_Metadata(t *testing.T) {
	d := NewFlatpak(DefaultRunner)
	if d.Name() != "flatpak" {
		t.Errorf("Name() = %q, want %q", d.Name(), "flatpak")
	}
	if d.NeedsSudo() {
		t.Error("NeedsSudo() = true, want false")
	}
}

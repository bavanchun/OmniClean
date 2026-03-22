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
			name:    "parses two apps with size",
			output:  "org.mozilla.firefox\t121.0\t286.4 MB\norg.videolan.VLC\t3.0.20\t56.6 MB\n",
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
				// 286.4 MB = 2864 * 1024 * 1024 / 10
				wantSize := int64(2864) * 1024 * 1024 / 10
				if pkgs[0].Size != wantSize {
					t.Errorf("Size = %d, want ~%d", pkgs[0].Size, wantSize)
				}
			},
		},
		{
			name:    "app with no version or size",
			output:  "org.example.App\n",
			wantLen: 1,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Version != "" {
					t.Errorf("Version = %q, want empty", pkgs[0].Version)
				}
				if pkgs[0].Size != 0 {
					t.Errorf("Size = %d, want 0", pkgs[0].Size)
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

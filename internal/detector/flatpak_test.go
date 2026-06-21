package detector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

// Flatpak is leaf-only: there is no documented read-only/dry-run form of
// `flatpak uninstall --unused`, so every listed app is treated as Manual and
// nothing is ever reported Orphan (trustworthy-or-silent).
func TestFlatpak_Classify_LeafOnly(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("list"), output: "org.mozilla.firefox\norg.videolan.VLC\n"},
		},
	}
	d := NewFlatpak(fr.run)
	d.stat = errStat()

	in := []pkg.Package{
		{Name: "org.mozilla.firefox", Manager: pkg.ManagerFlatpak},
		{Name: "org.videolan.VLC", Manager: pkg.ManagerFlatpak},
	}
	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	for _, p := range out {
		if p.Role != pkg.RoleManual {
			t.Errorf("%s role = %q, want manual (leaf-only)", p.Name, p.Role)
		}
		if p.Role == pkg.RoleOrphan {
			t.Errorf("%s must never be orphan: flatpak has no read-only orphan query", p.Name)
		}
	}

	// Only read-only commands; never a `flatpak uninstall`.
	for i := range fr.calls {
		if strings.Contains(fr.commandLine(i), "uninstall") {
			t.Errorf("classify issued mutating command: %q", fr.commandLine(i))
		}
	}
}

// TestFlatpak_Classify_RuntimesNeverSurfaced is the leaf-only regression guard:
// a runtime-style id that is NOT in the `flatpak list --app` set must stay
// RoleUnknown (never Manual, never Orphan), and apps stay Manual. Confirms
// flatpak never proposes a runtime as removable.
func TestFlatpak_Classify_RuntimesNeverSurfaced(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("list"), output: "org.mozilla.firefox\norg.videolan.VLC\ncom.spotify.Client\n"},
		},
	}
	d := NewFlatpak(fr.run)
	d.stat = errStat()

	in := []pkg.Package{
		{Name: "org.mozilla.firefox", Manager: pkg.ManagerFlatpak},      // app -> manual
		{Name: "org.freedesktop.Platform", Manager: pkg.ManagerFlatpak}, // runtime, not an app
	}
	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	roles := map[string]pkg.Role{}
	for _, p := range out {
		roles[p.Name] = p.Role
	}
	if roles["org.mozilla.firefox"] != pkg.RoleManual {
		t.Errorf("app role = %q, want manual", roles["org.mozilla.firefox"])
	}
	if roles["org.freedesktop.Platform"] != pkg.RoleUnknown {
		t.Errorf("runtime role = %q, want unknown (never removable)", roles["org.freedesktop.Platform"])
	}
	for _, p := range out {
		if p.Role == pkg.RoleOrphan {
			t.Errorf("%s marked orphan; flatpak must never surface orphans", p.Name)
		}
	}
}

func TestFlatpak_Classify_InstalledAtPresent(t *testing.T) {
	want := time.Date(2025, 3, 9, 0, 0, 0, 0, time.UTC)
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("list"), output: "org.mozilla.firefox\n"},
		},
	}
	d := NewFlatpak(fr.run)
	d.stat = fixedStat(want)

	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "org.mozilla.firefox", Manager: pkg.ManagerFlatpak}})
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if !out[0].InstalledAt.Equal(want) {
		t.Errorf("InstalledAt = %v, want %v", out[0].InstalledAt, want)
	}
}

func TestFlatpak_Classify_UnavailableDegradesToUnknown(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("list"), err: errors.New("flatpak: not found")},
		},
	}
	d := NewFlatpak(fr.run)
	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "org.mozilla.firefox", Manager: pkg.ManagerFlatpak}})
	if err != nil {
		t.Fatalf("Classify should degrade, not error: %v", err)
	}
	if out[0].Role != pkg.RoleUnknown {
		t.Errorf("role = %q, want unknown", out[0].Role)
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

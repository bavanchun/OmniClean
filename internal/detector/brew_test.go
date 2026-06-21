package detector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

func TestBrew_Classify(t *testing.T) {
	// leaves => manual (top-level): git, node
	// autoremove -n => orphan candidates: libfoo
	// linked-but-not-leaf (zlib) => dependency
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("leaves"), output: "git\nnode\n"},
			{match: argContains("autoremove"), output: "==> Would remove 1 formula:\nlibfoo\n"},
			{match: argContains("--cellar"), output: "/opt/homebrew/Cellar"},
		},
	}
	d := NewBrew(fr.run)
	d.stat = errStat() // install-time absent for this case

	in := []pkg.Package{
		{Name: "git", Manager: pkg.ManagerBrew},
		{Name: "libfoo", Manager: pkg.ManagerBrew},
		{Name: "zlib", Manager: pkg.ManagerBrew},
	}
	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}

	roles := map[string]pkg.Role{}
	for _, p := range out {
		roles[p.Name] = p.Role
	}
	if roles["git"] != pkg.RoleManual {
		t.Errorf("git role = %q, want manual", roles["git"])
	}
	if roles["libfoo"] != pkg.RoleOrphan {
		t.Errorf("libfoo role = %q, want orphan", roles["libfoo"])
	}
	if roles["zlib"] != pkg.RoleDependency {
		t.Errorf("zlib role = %q, want dependency", roles["zlib"])
	}

	// Only read-only commands may be issued. `brew uninstall` is always
	// mutating; `brew autoremove` is mutating UNLESS run with the -n dry-run flag.
	for i := range fr.calls {
		cl := fr.commandLine(i)
		mutating := strings.Contains(cl, "uninstall") ||
			(strings.Contains(cl, "autoremove") && !containsArg(fr.calls[i], "-n"))
		if mutating {
			t.Errorf("classify issued mutating command: %q", cl)
		}
	}
}

func TestBrew_Classify_InstalledAtPresent(t *testing.T) {
	want := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("leaves"), output: "git\n"},
			{match: argContains("autoremove"), output: ""},
			{match: argContains("--cellar"), output: "/opt/homebrew/Cellar"},
		},
	}
	d := NewBrew(fr.run)
	d.stat = fixedStat(want)

	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "git", Manager: pkg.ManagerBrew}})
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if !out[0].InstalledAt.Equal(want) {
		t.Errorf("InstalledAt = %v, want %v", out[0].InstalledAt, want)
	}
}

func TestBrew_Classify_UnavailableDegradesToUnknown(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("leaves"), err: errors.New("brew: not found")},
		},
	}
	d := NewBrew(fr.run)
	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "git", Manager: pkg.ManagerBrew}})
	if err != nil {
		t.Fatalf("Classify should degrade, not error: %v", err)
	}
	if out[0].Role != pkg.RoleUnknown {
		t.Errorf("role = %q, want unknown on unavailable query", out[0].Role)
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

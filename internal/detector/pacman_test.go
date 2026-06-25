package detector

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestPacman_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name: "parses installed packages",
			output: "acl 2.3.2-2\n" +
				"bash 5.3.15-1\n",
			wantLen: 2,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "acl" {
					t.Errorf("got Name %q, want %q", pkgs[0].Name, "acl")
				}
				if pkgs[0].Version != "2.3.2-2" {
					t.Errorf("got Version %q, want %q", pkgs[0].Version, "2.3.2-2")
				}
				if pkgs[0].Manager != pkg.ManagerType("pacman") {
					t.Errorf("got Manager %q, want %q", pkgs[0].Manager, "pacman")
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
			runErr:  errors.New("pacman: not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewPacman(runner)
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

func TestPacman_Classify_RealFixtures(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("-Qeq"), output: readFixture(t, "pacman-explicit-arch.txt")},
			{match: argContains("-Qtdq"), output: readFixture(t, "pacman-orphans-arch.txt")},
		},
	}
	d := NewPacman(fr.run)
	d.stat = errStat()

	in := []pkg.Package{
		{Name: "git", Version: "2.54.0-1", Manager: pkg.ManagerType("pacman")},            // manual in fixture
		{Name: "perl-error", Version: "0.17030-3", Manager: pkg.ManagerType("pacman")},  // orphan in fixture
		{Name: "ca-certificates", Version: "20240618-1", Manager: pkg.ManagerType("pacman")}, // dependency in fixture
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
		t.Errorf("git = %q, want manual", roles["git"])
	}
	if roles["perl-error"] != pkg.RoleOrphan {
		t.Errorf("perl-error = %q, want orphan", roles["perl-error"])
	}
	if roles["ca-certificates"] != pkg.RoleDependency {
		t.Errorf("ca-certificates = %q, want dependency", roles["ca-certificates"])
	}
}

func TestPacman_Classify_NoOrphansCleanExit(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("-Qeq"), output: "base\n"},
			{match: argContains("-Qtdq"), err: errors.New("exit status 1")}, // simulate no orphans
		},
	}
	d := NewPacman(fr.run)
	d.stat = errStat()

	in := []pkg.Package{
		{Name: "base", Manager: pkg.ManagerType("pacman")},
		{Name: "curl", Manager: pkg.ManagerType("pacman")},
	}
	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	roles := map[string]pkg.Role{}
	for _, p := range out {
		roles[p.Name] = p.Role
	}
	if roles["base"] != pkg.RoleManual {
		t.Errorf("base = %q, want manual", roles["base"])
	}
	if roles["curl"] != pkg.RoleDependency {
		t.Errorf("curl = %q, want dependency", roles["curl"])
	}
}

func TestPacman_Classify_UnavailableDegradesToUnknown(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("-Qeq"), err: errors.New("pacman: not found")},
		},
	}
	d := NewPacman(fr.run)
	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "git", Manager: pkg.ManagerType("pacman")}})
	if err != nil {
		t.Fatalf("Classify should degrade, not error: %v", err)
	}
	if out[0].Role != pkg.RoleUnknown {
		t.Errorf("role = %q, want unknown", out[0].Role)
	}
}

func TestPacman_Metadata(t *testing.T) {
	d := NewPacman(DefaultRunner)
	if d.Name() != "pacman" {
		t.Errorf("Name() = %q, want %q", d.Name(), "pacman")
	}
	if !d.NeedsSudo() {
		t.Error("NeedsSudo() = false, want true")
	}
	p := pkg.Package{Name: "git", Manager: pkg.ManagerType("pacman")}
	cmd := d.DryRunCommand(p)
	if !strings.Contains(cmd, "pacman") || !strings.Contains(cmd, "git") {
		t.Errorf("DryRunCommand() returned unexpected command: %q", cmd)
	}
}

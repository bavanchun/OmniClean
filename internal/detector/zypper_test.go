package detector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestZypper_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name: "parses rpm output",
			output: "curl\t7.88.1-10\t729088\t1779864471\tcommand line tool for transferring data with URL syntax\n" +
				"vim\t2:9.0.1378-2\t3805184\t1779864472\tVi IMproved - enhanced vi editor\n",
			wantLen: 2,
			check: func(t *testing.T, pkgs []pkg.Package) {
				t.Helper()
				if pkgs[0].Name != "curl" {
					t.Errorf("got Name %q, want %q", pkgs[0].Name, "curl")
				}
				if pkgs[0].Version != "7.88.1-10" {
					t.Errorf("got Version %q, want %q", pkgs[0].Version, "7.88.1-10")
				}
				if pkgs[0].Size != 729088 {
					t.Errorf("got Size %d, want %d", pkgs[0].Size, 729088)
				}
				if pkgs[0].InstalledAt.Unix() != 1779864471 {
					t.Errorf("got InstalledAt %v, want %v", pkgs[0].InstalledAt, time.Unix(1779864471, 0))
				}
				if pkgs[0].Manager != pkg.ManagerType("zypper") {
					t.Errorf("got Manager %q, want %q", pkgs[0].Manager, "zypper")
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
			runErr:  errors.New("rpm: not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewZypper(runner)
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

func TestZypper_Classify_RealFixtures(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("search"), output: readFixture(t, "zypper-installed-opensuse.txt")},
			{match: argContains("packages"), output: readFixture(t, "zypper-orphaned-opensuse.txt")},
		},
	}
	d := NewZypper(fr.run)

	in := []pkg.Package{
		{Name: "bash", Manager: pkg.ManagerType("zypper")},      // manual (i+) in fixture
		{Name: "git-core", Manager: pkg.ManagerType("zypper")},  // orphan (in unneeded table) in fixture
		{Name: "bash-sh", Manager: pkg.ManagerType("zypper")},   // dependency (i, not manual/orphan) in fixture
	}
	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	roles := map[string]pkg.Role{}
	for _, p := range out {
		roles[p.Name] = p.Role
	}
	if roles["bash"] != pkg.RoleManual {
		t.Errorf("bash = %q, want manual", roles["bash"])
	}
	if roles["git-core"] != pkg.RoleOrphan {
		t.Errorf("git-core = %q, want orphan", roles["git-core"])
	}
	if roles["bash-sh"] != pkg.RoleDependency {
		t.Errorf("bash-sh = %q, want dependency", roles["bash-sh"])
	}
}

func TestZypper_Classify_TableParser(t *testing.T) {
	input := `
S  | Name | Type | Version
---+------+------+--------
i+ | curl | pkg  | 7.88.1
i  | libc | pkg  | 2.36
`
	manual := parseZypperTable(input, true)
	if !manual["curl"] {
		t.Error("expected curl in manual set")
	}
	if manual["libc"] {
		t.Error("libc should not be in manual set")
	}

	all := parseZypperTable(input, false)
	if !all["curl"] || !all["libc"] {
		t.Errorf("expected curl and libc in all set, got %v", all)
	}
}

func TestZypper_Classify_UnavailableDegradesToUnknown(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("search"), err: errors.New("zypper: not found")},
		},
	}
	d := NewZypper(fr.run)
	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "bash", Manager: pkg.ManagerType("zypper")}})
	if err != nil {
		t.Fatalf("Classify should degrade, not error: %v", err)
	}
	if out[0].Role != pkg.RoleUnknown {
		t.Errorf("role = %q, want unknown", out[0].Role)
	}
}

func TestZypper_Metadata(t *testing.T) {
	d := NewZypper(DefaultRunner)
	if d.Name() != "zypper" {
		t.Errorf("Name() = %q, want %q", d.Name(), "zypper")
	}
	if !d.NeedsSudo() {
		t.Error("NeedsSudo() = false, want true")
	}
	p := pkg.Package{Name: "bash", Manager: pkg.ManagerType("zypper")}
	cmd := d.DryRunCommand(p)
	if !strings.Contains(cmd, "zypper") || !strings.Contains(cmd, "bash") {
		t.Errorf("DryRunCommand() returned unexpected command: %q", cmd)
	}
}

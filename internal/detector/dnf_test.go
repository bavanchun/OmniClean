package detector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestDNF_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
		check   func(t *testing.T, pkgs []pkg.Package)
	}{
		{
			name: "parses two installed rpm packages",
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
				if pkgs[0].Manager != pkg.ManagerType("dnf") {
					t.Errorf("got Manager %q, want %q", pkgs[0].Manager, "dnf")
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
			runErr:  errors.New("rpm: not found"),
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
			d := NewDNF(runner)
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

func TestDNF_Classify_RealFixtures(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("userinstalled"), output: readFixture(t, "dnf-userinstalled-fedora44.txt")},
			{match: argContains("unneeded"), output: readFixture(t, "dnf-unneeded-fedora44.txt")},
		},
	}
	d := NewDNF(fr.run)

	in := []pkg.Package{
		{Name: "bash", Manager: pkg.ManagerType("dnf")},      // manual in fixture
		{Name: "dbus-libs", Manager: pkg.ManagerType("dnf")}, // orphan in fixture
		{Name: "openssl", Manager: pkg.ManagerType("dnf")},   // dependency in fixture
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
	if roles["dbus-libs"] != pkg.RoleOrphan {
		t.Errorf("dbus-libs = %q, want orphan", roles["dbus-libs"])
	}
	if roles["openssl"] != pkg.RoleDependency {
		t.Errorf("openssl = %q, want dependency", roles["openssl"])
	}
}

func TestDNF_Classify_UnavailableDegradesToUnknown(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("userinstalled"), err: errors.New("dnf: not found")},
		},
	}
	d := NewDNF(fr.run)
	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "bash", Manager: pkg.ManagerType("dnf")}})
	if err != nil {
		t.Fatalf("Classify should degrade, not error: %v", err)
	}
	if out[0].Role != pkg.RoleUnknown {
		t.Errorf("role = %q, want unknown", out[0].Role)
	}
}

func TestDNF_Metadata(t *testing.T) {
	d := NewDNF(DefaultRunner)
	if d.Name() != "dnf" {
		t.Errorf("Name() = %q, want %q", d.Name(), "dnf")
	}
	if !d.NeedsSudo() {
		t.Error("NeedsSudo() = false, want true")
	}
	p := pkg.Package{Name: "bash", Manager: pkg.ManagerType("dnf")}
	cmd := d.DryRunCommand(p)
	if !strings.Contains(cmd, "remove") || !strings.Contains(cmd, "bash") {
		t.Errorf("DryRunCommand() returned unexpected command: %q", cmd)
	}
}

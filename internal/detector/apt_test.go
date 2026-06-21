package detector

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

func TestAPT_Classify(t *testing.T) {
	// apt-mark showmanual => manual: curl, vim
	// apt-get autoremove --dry-run => "Remv libfoo ..." => orphan: libfoo
	// libbar installed but neither => dependency
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("showmanual"), output: "curl\nvim\n"},
			{match: argContains("autoremove"), output: "Reading package lists...\n" +
				"The following packages will be REMOVED:\n  libfoo\n" +
				"Remv libfoo [1.2-3]\n"},
		},
	}
	d := NewAPT(fr.run)
	d.stat = errStat()

	in := []pkg.Package{
		{Name: "curl", Manager: pkg.ManagerAPT},
		{Name: "libfoo", Manager: pkg.ManagerAPT},
		{Name: "libbar", Manager: pkg.ManagerAPT},
	}
	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	roles := map[string]pkg.Role{}
	for _, p := range out {
		roles[p.Name] = p.Role
	}
	if roles["curl"] != pkg.RoleManual {
		t.Errorf("curl role = %q, want manual", roles["curl"])
	}
	if roles["libfoo"] != pkg.RoleOrphan {
		t.Errorf("libfoo role = %q, want orphan", roles["libfoo"])
	}
	if roles["libbar"] != pkg.RoleDependency {
		t.Errorf("libbar role = %q, want dependency", roles["libbar"])
	}

	// Classify must stay read-only: the autoremove probe must carry --dry-run,
	// and no `apt-get remove`/`sudo` invocation may appear.
	for i := range fr.calls {
		cl := fr.commandLine(i)
		if strings.Contains(cl, "sudo") {
			t.Errorf("classify escalated privileges: %q", cl)
		}
		if strings.Contains(cl, "autoremove") && !containsArg(fr.calls[i], "--dry-run") {
			t.Errorf("autoremove probe not dry-run: %q", cl)
		}
	}
}

func TestAPT_Classify_InstalledAtPresent(t *testing.T) {
	want := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("showmanual"), output: "curl\n"},
			{match: argContains("autoremove"), output: ""},
		},
	}
	d := NewAPT(fr.run)
	d.stat = fixedStat(want)

	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "curl", Manager: pkg.ManagerAPT}})
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if !out[0].InstalledAt.Equal(want) {
		t.Errorf("InstalledAt = %v, want %v", out[0].InstalledAt, want)
	}
}

func TestAPT_Classify_UnavailableDegradesToUnknown(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{match: argContains("showmanual"), err: errors.New("apt-mark: not found")},
		},
	}
	d := NewAPT(fr.run)
	out, err := d.Classify(context.Background(), []pkg.Package{{Name: "curl", Manager: pkg.ManagerAPT}})
	if err != nil {
		t.Fatalf("Classify should degrade, not error: %v", err)
	}
	if out[0].Role != pkg.RoleUnknown {
		t.Errorf("role = %q, want unknown", out[0].Role)
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

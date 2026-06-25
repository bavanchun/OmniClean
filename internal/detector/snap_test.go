package detector

import (
	"context"
	"errors"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestSnap_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
	}{
		{
			name: "parses two packages",
			output: "Name    Version   Rev  Tracking       Publisher   Notes\n" +
				"firefox 121.0     3000 latest/stable  mozilla     -\n" +
				"vlc     3.0.20    3379 latest/stable  videolan    -\n",
			wantLen: 2,
		},
		{
			name:    "empty output returns empty (no panic)",
			output:  "",
			wantLen: 0,
		},
		{
			name:    "header only returns empty (no panic on lines[1:])",
			output:  "Name    Version   Rev  Tracking       Publisher   Notes",
			wantLen: 0,
		},
		{
			name:    "command error propagates",
			runErr:  errors.New("snap: command not found"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewSnap(runner)
			pkgs, err := d.ListPackages(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("ListPackages() error = %v, wantErr %v", err, tc.wantErr)
			}
			if len(pkgs) != tc.wantLen {
				t.Errorf("got %d packages, want %d", len(pkgs), tc.wantLen)
			}
			for _, p := range pkgs {
				if p.Manager != pkg.ManagerSnap {
					t.Errorf("Manager = %q, want %q", p.Manager, pkg.ManagerSnap)
				}
			}
		})
	}
}

func TestSnap_Metadata(t *testing.T) {
	d := NewSnap(DefaultRunner)
	if d.Name() != "snap" {
		t.Errorf("Name() = %q, want %q", d.Name(), "snap")
	}
	if !d.NeedsSudo() {
		t.Error("NeedsSudo() = false, want true")
	}
}

func TestSnap_Classify(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{
				match: func(_ string, args []string) bool { return len(args) > 0 && args[0] == "list" },
				output: "Name    Version   Rev  Tracking       Publisher   Notes\n" +
					"firefox 121.0     3000 latest/stable  mozilla     -\n" +
					"vlc     3.0.20    3379 latest/stable  videolan    -\n",
			},
		},
	}
	d := NewSnap(fr.run)

	in := []pkg.Package{
		{Name: "firefox", Manager: pkg.ManagerSnap},
		{Name: "vlc", Manager: pkg.ManagerSnap},
		{Name: "uninstalled-snap", Manager: pkg.ManagerSnap},
	}

	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}

	roles := map[string]pkg.Role{}
	for _, p := range out {
		roles[p.Name] = p.Role
	}

	if roles["firefox"] != pkg.RoleManual {
		t.Errorf("firefox role = %q, want manual", roles["firefox"])
	}
	if roles["vlc"] != pkg.RoleManual {
		t.Errorf("vlc role = %q, want manual", roles["vlc"])
	}
	if roles["uninstalled-snap"] != pkg.RoleUnknown {
		t.Errorf("uninstalled-snap role = %q, want unknown", roles["uninstalled-snap"])
	}
}

func TestSnap_Classify_UnavailableDegradesToUnknown(t *testing.T) {
	fr := &fakeRunner{
		responses: []fakeResponse{
			{
				match: func(_ string, args []string) bool { return len(args) > 0 && args[0] == "list" },
				err:   errors.New("snap command failed"),
			},
		},
	}
	d := NewSnap(fr.run)

	in := []pkg.Package{
		{Name: "firefox", Manager: pkg.ManagerSnap},
	}

	out, err := d.Classify(context.Background(), in)
	if err != nil {
		t.Fatalf("Classify should degrade, not error: %v", err)
	}

	if out[0].Role != pkg.RoleUnknown {
		t.Errorf("firefox role = %q, want unknown", out[0].Role)
	}
}

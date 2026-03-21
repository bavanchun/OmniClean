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

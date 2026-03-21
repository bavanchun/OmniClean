package detector

import (
	"context"
	"errors"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestNPM_ListPackages(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantLen int
		wantErr bool
	}{
		{
			name: "parses valid JSON with two packages",
			output: `{
				"dependencies": {
					"typescript": {"version": "5.3.3"},
					"eslint":     {"version": "8.56.0"}
				}
			}`,
			wantLen: 2,
		},
		{
			name:    "empty dependencies",
			output:  `{"dependencies": {}}`,
			wantLen: 0,
		},
		{
			name:    "invalid JSON with no output returns error",
			output:  "",
			runErr:  errors.New("npm exit 1"),
			wantErr: true,
		},
		{
			name:    "invalid JSON alone returns error",
			output:  "not json",
			wantErr: true,
		},
		{
			name:    "npm exit 1 with valid JSON (peer dep warnings) still parses",
			output:  `{"dependencies": {"typescript": {"version": "5.3.3"}}}`,
			runErr:  errors.New("npm: exit status 1"),
			wantLen: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := func(_ context.Context, _ string, _ ...string) (string, error) {
				return tc.output, tc.runErr
			}
			d := NewNPM(runner)
			pkgs, err := d.ListPackages(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("ListPackages() error = %v, wantErr %v", err, tc.wantErr)
			}
			if len(pkgs) != tc.wantLen {
				t.Errorf("got %d packages, want %d", len(pkgs), tc.wantLen)
			}
			for _, p := range pkgs {
				if p.Manager != pkg.ManagerNPM {
					t.Errorf("Manager = %q, want %q", p.Manager, pkg.ManagerNPM)
				}
			}
		})
	}
}

func TestNPM_Metadata(t *testing.T) {
	d := NewNPM(DefaultRunner)
	if d.Name() != "npm" {
		t.Errorf("Name() = %q, want %q", d.Name(), "npm")
	}
	if d.NeedsSudo() {
		t.Error("NeedsSudo() = true, want false")
	}
}

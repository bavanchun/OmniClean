package cleaner

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// mockDetector is a minimal Detector for testing.
type mockDetector struct {
	name         string
	needsSudo    bool
	uninstallErr error
	dryRunCmd    string
}

func (m *mockDetector) Name() string                                          { return m.name }
func (m *mockDetector) Available() bool                                       { return true }
func (m *mockDetector) NeedsSudo() bool                                       { return m.needsSudo }
func (m *mockDetector) DryRunCommand(p pkg.Package) string                    { return m.dryRunCmd }
func (m *mockDetector) UninstallExecCmd(_ pkg.Package) *exec.Cmd              { return nil }
func (m *mockDetector) ListPackages(_ context.Context) ([]pkg.Package, error) { return nil, nil }
func (m *mockDetector) Uninstall(_ context.Context, _ pkg.Package) error {
	return m.uninstallErr
}

var _ detector.Detector = (*mockDetector)(nil)

func TestCleaner_Uninstall_Success(t *testing.T) {
	d := &mockDetector{name: "brew", dryRunCmd: "brew uninstall git"}
	c := New([]detector.Detector{d})
	p := pkg.Package{Name: "git", Manager: "brew"}

	results := c.Uninstall(context.Background(), []pkg.Package{p}, false)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}
	if results[0].DryRunCmd != "" {
		t.Errorf("DryRunCmd should be empty on real uninstall, got %q", results[0].DryRunCmd)
	}
}

func TestCleaner_Uninstall_Error(t *testing.T) {
	d := &mockDetector{name: "pip", uninstallErr: errors.New("permission denied")}
	c := New([]detector.Detector{d})
	p := pkg.Package{Name: "requests", Manager: "pip"}

	results := c.Uninstall(context.Background(), []pkg.Package{p}, false)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCleaner_Uninstall_DryRun(t *testing.T) {
	d := &mockDetector{name: "cargo", dryRunCmd: "cargo uninstall ripgrep"}
	c := New([]detector.Detector{d})
	p := pkg.Package{Name: "ripgrep", Manager: "cargo"}

	results := c.Uninstall(context.Background(), []pkg.Package{p}, true)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("dry-run should not set Err: %v", results[0].Err)
	}
	if results[0].DryRunCmd != "cargo uninstall ripgrep" {
		t.Errorf("DryRunCmd = %q, want %q", results[0].DryRunCmd, "cargo uninstall ripgrep")
	}
}

func TestCleaner_Uninstall_MissingDetector(t *testing.T) {
	c := New([]detector.Detector{})
	p := pkg.Package{Name: "vim", Manager: "apt"}

	results := c.Uninstall(context.Background(), []pkg.Package{p}, false)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected error for missing detector, got nil")
	}
}

func TestCleaner_Uninstall_MultiplePackages(t *testing.T) {
	d1 := &mockDetector{name: "brew"}
	d2 := &mockDetector{name: "pip", uninstallErr: errors.New("fail")}
	c := New([]detector.Detector{d1, d2})
	packages := []pkg.Package{
		{Name: "git", Manager: "brew"},
		{Name: "requests", Manager: "pip"},
	}

	results := c.Uninstall(context.Background(), packages, false)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("brew result: unexpected error: %v", results[0].Err)
	}
	if results[1].Err == nil {
		t.Error("pip result: expected error, got nil")
	}
}

func TestFindLeftovers_NonExistentPackage(t *testing.T) {
	// A package name that certainly has no leftover dirs.
	leftovers := FindLeftovers("zzz-nonexistent-package-omniclean-test")
	if len(leftovers) != 0 {
		t.Errorf("expected no leftovers, got %v", leftovers)
	}
}

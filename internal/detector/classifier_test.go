package detector

import (
	"context"
	"os/exec"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// nonClassifierDetector implements Detector but NOT RemovableClassifier.
type nonClassifierDetector struct{}

func (nonClassifierDetector) Name() string                           { return "noclass" }
func (nonClassifierDetector) Available() bool                        { return true }
func (nonClassifierDetector) NeedsSudo() bool                        { return false }
func (nonClassifierDetector) DryRunCommand(pkg.Package) string       { return "" }
func (nonClassifierDetector) UninstallExecCmd(pkg.Package) *exec.Cmd { return nil }
func (nonClassifierDetector) ListPackages(context.Context) ([]pkg.Package, error) {
	return nil, nil
}
func (nonClassifierDetector) Uninstall(context.Context, pkg.Package) error { return nil }

// classifierDetector also implements RemovableClassifier and records the call.
type classifierDetector struct {
	nonClassifierDetector
	called bool
}

func (c *classifierDetector) Classify(_ context.Context, pkgs []pkg.Package) ([]pkg.Package, error) {
	c.called = true
	out := make([]pkg.Package, len(pkgs))
	copy(out, pkgs)
	for i := range out {
		out[i].Role = pkg.RoleManual
	}
	return out, nil
}

func TestClassifyIfSupported_NonClassifierUnchanged(t *testing.T) {
	in := []pkg.Package{{Name: "git", Manager: pkg.ManagerBrew}}
	out, err := ClassifyIfSupported(context.Background(), nonClassifierDetector{}, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Role != pkg.RoleUnknown {
		t.Errorf("non-classifier should leave Role unknown, got %q", out[0].Role)
	}
}

func TestClassifyIfSupported_InvokesClassifier(t *testing.T) {
	d := &classifierDetector{}
	in := []pkg.Package{{Name: "git", Manager: pkg.ManagerBrew}}
	out, err := ClassifyIfSupported(context.Background(), d, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.called {
		t.Error("Classify was not invoked on a detector implementing RemovableClassifier")
	}
	if out[0].Role != pkg.RoleManual {
		t.Errorf("classifier output Role = %q, want RoleManual", out[0].Role)
	}
}

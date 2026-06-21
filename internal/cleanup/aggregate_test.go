package cleanup

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// baseDetector implements detector.Detector with no-op behavior; concrete
// fakes embed it and override ListPackages / add Classify as needed.
type baseDetector struct{ name string }

func (b baseDetector) Name() string                           { return b.name }
func (b baseDetector) Available() bool                        { return true }
func (b baseDetector) NeedsSudo() bool                        { return false }
func (b baseDetector) DryRunCommand(pkg.Package) string       { return "" }
func (b baseDetector) UninstallExecCmd(pkg.Package) *exec.Cmd { return nil }
func (b baseDetector) ListPackages(context.Context) ([]pkg.Package, error) {
	return nil, nil
}
func (b baseDetector) Uninstall(context.Context, pkg.Package) error { return nil }

// classifierFake returns a fixed classified package set.
type classifierFake struct {
	baseDetector
	out []pkg.Package
}

func (c classifierFake) ListPackages(context.Context) ([]pkg.Package, error) {
	// Return names only; Classify assigns roles.
	bare := make([]pkg.Package, len(c.out))
	for i, p := range c.out {
		bare[i] = pkg.Package{Name: p.Name, Manager: p.Manager}
	}
	return bare, nil
}

func (c classifierFake) Classify(_ context.Context, _ []pkg.Package) ([]pkg.Package, error) {
	return c.out, nil
}

// nonClassifierFake lists packages but never implements RemovableClassifier,
// so its packages keep RoleUnknown and must be excluded from suggestions.
type nonClassifierFake struct {
	baseDetector
	out []pkg.Package
}

func (n nonClassifierFake) ListPackages(context.Context) ([]pkg.Package, error) {
	return n.out, nil
}

// slowDetector blocks in ListPackages until the context is cancelled, modeling
// a hung manager query that must degrade gracefully (be skipped).
type slowDetector struct{ baseDetector }

func (s slowDetector) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestAggregate_FiltersAndSorts(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	clf := classifierFake{
		baseDetector: baseDetector{name: "brew"},
		out: []pkg.Package{
			{Name: "manual-new", Manager: pkg.ManagerBrew, Role: pkg.RoleManual, InstalledAt: t1},
			{Name: "manual-old", Manager: pkg.ManagerBrew, Role: pkg.RoleManual, InstalledAt: t0},
			{Name: "orphan-noinstalltime", Manager: pkg.ManagerBrew, Role: pkg.RoleOrphan},
			{Name: "dep-hidden", Manager: pkg.ManagerBrew, Role: pkg.RoleDependency},
		},
	}
	non := nonClassifierFake{
		baseDetector: baseDetector{name: "npm"},
		out:          []pkg.Package{{Name: "left-pad", Manager: pkg.ManagerNPM}}, // RoleUnknown
	}

	got := Aggregate(context.Background(), []detector.Detector{clf, non})

	// Dependency and Unknown must never appear.
	for _, p := range got {
		if p.Role == pkg.RoleDependency || p.Role == pkg.RoleUnknown {
			t.Errorf("removable list contains non-candidate %s (role %q)", p.Name, p.Role)
		}
		if p.Name == "left-pad" {
			t.Errorf("Unknown package from non-classifier must be excluded")
		}
	}

	wantOrder := []string{"orphan-noinstalltime", "manual-old", "manual-new"}
	if len(got) != len(wantOrder) {
		t.Fatalf("got %d candidates, want %d: %+v", len(got), len(wantOrder), names(got))
	}
	for i, w := range wantOrder {
		if got[i].Name != w {
			t.Errorf("order[%d] = %s, want %s (full: %v)", i, got[i].Name, w, names(got))
		}
	}
}

func TestAggregate_SlowDetectorDegrades(t *testing.T) {
	clf := classifierFake{
		baseDetector: baseDetector{name: "brew"},
		out:          []pkg.Package{{Name: "git", Manager: pkg.ManagerBrew, Role: pkg.RoleManual}},
	}
	slow := slowDetector{baseDetector{name: "apt"}}

	done := make(chan []pkg.Package, 1)
	go func() {
		done <- aggregate(context.Background(), []detector.Detector{clf, slow}, 30*time.Millisecond)
	}()

	select {
	case got := <-done:
		if len(got) != 1 || got[0].Name != "git" {
			t.Errorf("slow detector should be skipped, got %v", names(got))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Aggregate hung on slow detector instead of degrading via timeout")
	}
}

func names(pkgs []pkg.Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.Name
	}
	return out
}

package cleanup

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// fakeDetector is a minimal non-sudo detector for TUI wiring tests.
type fakeDetector struct{ name string }

func (f fakeDetector) Name() string                           { return f.name }
func (f fakeDetector) Available() bool                        { return true }
func (f fakeDetector) NeedsSudo() bool                        { return false }
func (f fakeDetector) DryRunCommand(p pkg.Package) string     { return f.name + " uninstall " + p.Name }
func (f fakeDetector) UninstallExecCmd(pkg.Package) *exec.Cmd { return nil }
func (f fakeDetector) ListPackages(context.Context) ([]pkg.Package, error) {
	return nil, nil
}
func (f fakeDetector) Uninstall(context.Context, pkg.Package) error { return nil }

func newTestApp() *App {
	a := New(Config{Detectors: []detector.Detector{fakeDetector{name: "brew"}}, DryRun: true})
	a.width, a.height = 100, 30
	return a
}

func TestCleanupTUI_LoadAndSelectFlow(t *testing.T) {
	a := newTestApp()
	cands := []pkg.Package{
		{Name: "orphan-lib", Manager: pkg.ManagerBrew, Role: pkg.RoleOrphan},
		{Name: "leaf-tool", Manager: pkg.ManagerBrew, Role: pkg.RoleManual, InstalledAt: time.Now()},
	}
	a.Update(loadDoneMsg{candidates: cands})
	if a.state != stateList {
		t.Fatalf("after load, state = %v, want stateList", a.state)
	}

	// Toggle the focused candidate, then confirm.
	a.handleKey("space")
	if got := a.selectedPkgs(); len(got) != 1 || got[0].Name != "orphan-lib" {
		t.Fatalf("space did not select focused candidate: %+v", got)
	}
	a.handleKey("enter")
	if a.state != stateConfirm {
		t.Fatalf("enter on selection -> state %v, want stateConfirm", a.state)
	}

	// Backing out returns to the list.
	a.handleKey("n")
	if a.state != stateList {
		t.Errorf("'n' on confirm -> state %v, want stateList", a.state)
	}
}

func TestCleanupTUI_RoleBadgesRender(t *testing.T) {
	a := newTestApp()
	a.Update(loadDoneMsg{candidates: []pkg.Package{
		{Name: "orphan-lib", Manager: pkg.ManagerBrew, Role: pkg.RoleOrphan},
		{Name: "leaf-tool", Manager: pkg.ManagerBrew, Role: pkg.RoleManual},
	}})
	out := a.viewList()
	if !strings.Contains(out, "orphan") {
		t.Errorf("viewList missing 'orphan' badge\n%s", out)
	}
	if !strings.Contains(out, "leaf") {
		t.Errorf("viewList missing 'leaf' badge\n%s", out)
	}
	if !strings.Contains(out, "orphan-lib") || !strings.Contains(out, "leaf-tool") {
		t.Errorf("viewList missing candidate names")
	}
}

func TestCleanupTUI_EmptyListMessage(t *testing.T) {
	a := newTestApp()
	a.Update(loadDoneMsg{candidates: nil})
	if !strings.Contains(a.viewList(), "Nothing to clean up") {
		t.Errorf("empty list should show a friendly message")
	}
}

func TestRoleBadgeAndInstalledAtLabel(t *testing.T) {
	if roleBadge(pkg.RoleOrphan) != "orphan" {
		t.Errorf("roleBadge(orphan) = %q", roleBadge(pkg.RoleOrphan))
	}
	if roleBadge(pkg.RoleManual) != "leaf" {
		t.Errorf("roleBadge(manual) = %q", roleBadge(pkg.RoleManual))
	}
	if got := installedAtLabel(pkg.Package{}); got != "—" {
		t.Errorf("installedAtLabel(zero) = %q, want em dash", got)
	}
}

package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Snap detects packages installed via Snap.
type Snap struct {
	run CommandRunner
}

// NewSnap creates a Snap detector with the given command runner.
func NewSnap(run CommandRunner) *Snap {
	return &Snap{run: run}
}

func (s *Snap) Name() string    { return "snap" }
func (s *Snap) NeedsSudo() bool { return true }
func (s *Snap) Available() bool { return LookPath("snap") }

func (s *Snap) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("sudo snap remove %s", p.Name)
}

func (s *Snap) UninstallExecCmd(p pkg.Package) *exec.Cmd {
	return NewSudoExecCmd("sudo", "snap", "remove", p.Name)
}

func (s *Snap) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := s.run(ctx, "snap", "list")
	if err != nil {
		return nil, fmt.Errorf("snap list packages: %w", err)
	}

	var packages []pkg.Package
	lines := strings.Split(out, "\n")
	// Skip header line; guard against empty output.
	if len(lines) <= 1 {
		return packages, nil
	}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		packages = append(packages, pkg.Package{
			Name:    fields[0],
			Version: fields[1],
			Manager: pkg.ManagerSnap,
		})
	}
	return packages, nil
}

// Uninstall is not used for Snap since NeedsSudo=true.
func (s *Snap) Uninstall(_ context.Context, p pkg.Package) error {
	return fmt.Errorf("snap: use UninstallExecCmd for interactive sudo uninstall of %s", p.Name)
}

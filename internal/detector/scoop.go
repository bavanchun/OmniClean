package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Scoop detects packages installed via Scoop.
type Scoop struct {
	run CommandRunner
}

// NewScoop creates a Scoop detector with the given command runner.
func NewScoop(run CommandRunner) *Scoop {
	return &Scoop{run: run}
}

func (s *Scoop) Name() string    { return "scoop" }
func (s *Scoop) NeedsSudo() bool { return false }
func (s *Scoop) Available() bool { return LookPath("scoop") }

func (s *Scoop) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("scoop uninstall %s", p.Name)
}

func (s *Scoop) UninstallExecCmd(_ pkg.Package) *exec.Cmd { return nil }

func (s *Scoop) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	// scoop list outputs a table: Name  Version  Source  Updated  Info
	out, err := s.run(ctx, "scoop", "list")
	if err != nil {
		return nil, fmt.Errorf("scoop list packages: %w", err)
	}

	var packages []pkg.Package
	lines := strings.Split(out, "\n")
	headerPassed := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "---") {
			headerPassed = true
			continue
		}
		if !headerPassed || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		packages = append(packages, pkg.Package{
			Name:    fields[0],
			Version: fields[1],
			Manager: pkg.ManagerScoop,
		})
	}
	return packages, nil
}

func (s *Scoop) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := s.run(ctx, "scoop", "uninstall", p.Name)
	if err != nil {
		return fmt.Errorf("scoop uninstall %s: %w", p.Name, err)
	}
	return nil
}

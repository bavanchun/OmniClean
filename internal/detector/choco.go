package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Choco detects packages installed via Chocolatey.
type Choco struct {
	run CommandRunner
}

// NewChoco creates a Choco detector with the given command runner.
func NewChoco(run CommandRunner) *Choco {
	return &Choco{run: run}
}

func (c *Choco) Name() string    { return "choco" }
func (c *Choco) NeedsSudo() bool { return false }
func (c *Choco) Available() bool { return LookPath("choco") }

func (c *Choco) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("choco uninstall %s -y", p.Name)
}

func (c *Choco) UninstallExecCmd(_ pkg.Package) *exec.Cmd { return nil }

func (c *Choco) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	// choco list --local-only outputs: "name version"
	out, err := c.run(ctx, "choco", "list", "--local-only")
	if err != nil {
		return nil, fmt.Errorf("choco list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "packages installed") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		packages = append(packages, pkg.Package{
			Name:    parts[0],
			Version: parts[1],
			Manager: pkg.ManagerChoco,
		})
	}
	return packages, nil
}

func (c *Choco) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := c.run(ctx, "choco", "uninstall", p.Name, "-y")
	if err != nil {
		return fmt.Errorf("choco uninstall %s: %w", p.Name, err)
	}
	return nil
}

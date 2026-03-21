package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Brew detects packages installed via Homebrew.
type Brew struct {
	run CommandRunner
}

// NewBrew creates a Brew detector with the given command runner.
func NewBrew(run CommandRunner) *Brew {
	return &Brew{run: run}
}

func (b *Brew) Name() string    { return "brew" }
func (b *Brew) NeedsSudo() bool { return false }
func (b *Brew) Available() bool { return LookPath("brew") }

func (b *Brew) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("brew uninstall %s", p.Name)
}

func (b *Brew) UninstallExecCmd(_ pkg.Package) *exec.Cmd { return nil }

func (b *Brew) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := b.run(ctx, "brew", "list", "--versions")
	if err != nil {
		return nil, fmt.Errorf("brew list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		p := pkg.Package{
			Name:    parts[0],
			Manager: pkg.ManagerBrew,
		}
		if len(parts) >= 2 {
			p.Version = parts[len(parts)-1]
		}
		packages = append(packages, p)
	}
	return packages, nil
}

func (b *Brew) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := b.run(ctx, "brew", "uninstall", p.Name)
	if err != nil {
		return fmt.Errorf("brew uninstall %s: %w", p.Name, err)
	}
	return nil
}

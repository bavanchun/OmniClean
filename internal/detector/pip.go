package detector

import (
	"context"
	"fmt"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Pip detects Python packages installed via pip.
type Pip struct {
	run CommandRunner
}

// NewPip creates a Pip detector with the given command runner.
func NewPip(run CommandRunner) *Pip {
	return &Pip{run: run}
}

func (p *Pip) Name() string { return "pip" }

func (p *Pip) Available() bool {
	return LookPath("pip3") || LookPath("pip")
}

func (p *Pip) binary() string {
	if LookPath("pip3") {
		return "pip3"
	}
	return "pip"
}

func (p *Pip) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	// freeze format: name==version
	out, err := p.run(ctx, p.binary(), "list", "--format=freeze")
	if err != nil {
		return nil, fmt.Errorf("pip list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "==", 2)
		pk := pkg.Package{
			Name:    parts[0],
			Manager: pkg.ManagerPip,
		}
		if len(parts) >= 2 {
			pk.Version = parts[1]
		}
		packages = append(packages, pk)
	}
	return packages, nil
}

func (p *Pip) Uninstall(ctx context.Context, pk pkg.Package, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] %s uninstall -y %s\n", p.binary(), pk.Name)
		return nil
	}
	_, err := p.run(ctx, p.binary(), "uninstall", "-y", pk.Name)
	if err != nil {
		return fmt.Errorf("pip uninstall %s: %w", pk.Name, err)
	}
	return nil
}

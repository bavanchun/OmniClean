package detector

import (
	"context"
	"fmt"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Cargo detects packages installed via cargo install.
type Cargo struct {
	run CommandRunner
}

// NewCargo creates a Cargo detector with the given command runner.
func NewCargo(run CommandRunner) *Cargo {
	return &Cargo{run: run}
}

func (c *Cargo) Name() string { return "cargo" }

func (c *Cargo) Available() bool {
	return LookPath("cargo")
}

func (c *Cargo) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	// cargo install --list outputs lines like: "crate v1.2.3 (...):"
	out, err := c.run(ctx, "cargo", "install", "--list")
	if err != nil {
		return nil, fmt.Errorf("cargo list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		// Installed binaries are indented; skip them
		if strings.HasPrefix(line, " ") || strings.TrimSpace(line) == "" {
			continue
		}
		// Format: "name v1.2.3:"  or  "name v1.2.3 (path ...):"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		version := strings.TrimSuffix(strings.TrimPrefix(parts[1], "v"), ":")
		packages = append(packages, pkg.Package{
			Name:    name,
			Version: version,
			Manager: pkg.ManagerCargo,
		})
	}
	return packages, nil
}

func (c *Cargo) Uninstall(ctx context.Context, p pkg.Package, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] cargo uninstall %s\n", p.Name)
		return nil
	}
	_, err := c.run(ctx, "cargo", "uninstall", p.Name)
	if err != nil {
		return fmt.Errorf("cargo uninstall %s: %w", p.Name, err)
	}
	return nil
}

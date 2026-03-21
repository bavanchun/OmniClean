package detector

import (
	"context"
	"fmt"
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

func (s *Snap) Name() string { return "snap" }

func (s *Snap) Available() bool {
	return LookPath("snap")
}

func (s *Snap) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := s.run(ctx, "snap", "list")
	if err != nil {
		return nil, fmt.Errorf("snap list packages: %w", err)
	}

	var packages []pkg.Package
	lines := strings.Split(out, "\n")
	// Skip header line
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

func (s *Snap) Uninstall(ctx context.Context, p pkg.Package, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] sudo snap remove %s\n", p.Name)
		return nil
	}
	_, err := s.run(ctx, "sudo", "snap", "remove", p.Name)
	if err != nil {
		return fmt.Errorf("snap uninstall %s: %w", p.Name, err)
	}
	return nil
}

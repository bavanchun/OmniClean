// Package cleaner orchestrates package uninstallation and leftover file detection.
package cleaner

import (
	"context"
	"os"
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Cleaner holds a map of detectors by manager name for fast lookup.
type Cleaner struct {
	detectors map[string]detector.Detector
}

// New creates a Cleaner from a slice of detectors.
func New(detectors []detector.Detector) *Cleaner {
	m := make(map[string]detector.Detector, len(detectors))
	for _, d := range detectors {
		m[d.Name()] = d
	}
	return &Cleaner{detectors: m}
}

// Uninstall removes each package using its associated detector.
// It returns one result per package regardless of success or failure.
func (c *Cleaner) Uninstall(ctx context.Context, packages []pkg.Package, dryRun bool) []pkg.UninstallResult {
	results := make([]pkg.UninstallResult, 0, len(packages))
	for _, p := range packages {
		d, ok := c.detectors[string(p.Manager)]
		if !ok {
			results = append(results, pkg.UninstallResult{Package: p})
			continue
		}
		err := d.Uninstall(ctx, p, dryRun)
		results = append(results, pkg.UninstallResult{Package: p, Err: err})
	}
	return results
}

// FindLeftovers returns paths of leftover config/cache/data directories for
// a given package name. It checks the most common XDG locations.
func FindLeftovers(pkgName string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".config", pkgName),
		filepath.Join(home, ".local", "share", pkgName),
		filepath.Join(home, ".cache", pkgName),
		filepath.Join(home, "."+pkgName),
		filepath.Join("/etc", pkgName),
		filepath.Join("/var", "lib", pkgName),
	}

	var found []string
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
	}
	return found
}

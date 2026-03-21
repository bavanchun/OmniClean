// Package cleaner orchestrates package uninstallation and leftover file detection.
package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
// When dryRun is true, it populates DryRunCmd instead of executing.
// It returns one result per package regardless of success or failure.
func (c *Cleaner) Uninstall(ctx context.Context, packages []pkg.Package, dryRun bool) []pkg.UninstallResult {
	results := make([]pkg.UninstallResult, 0, len(packages))
	for _, p := range packages {
		d, ok := c.detectors[string(p.Manager)]
		if !ok {
			results = append(results, pkg.UninstallResult{
				Package: p,
				Err:     fmt.Errorf("no detector available for manager %q", p.Manager),
			})
			continue
		}

		result := pkg.UninstallResult{Package: p}

		if dryRun {
			result.DryRunCmd = d.DryRunCommand(p)
		} else {
			result.Err = d.Uninstall(ctx, p)
			if result.Err == nil {
				result.Leftovers = FindLeftovers(p.Name)
			}
		}

		results = append(results, result)
	}
	return results
}

// FindLeftovers returns paths of leftover config/cache/data directories for
// a given package name, using platform-appropriate locations.
func FindLeftovers(pkgName string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	configDir, _ := os.UserConfigDir()
	cacheDir, _ := os.UserCacheDir()

	candidates := []string{
		filepath.Join(home, "."+pkgName), // legacy dot-dir (all platforms)
	}

	if configDir != "" {
		candidates = append(candidates, filepath.Join(configDir, pkgName))
	}
	if cacheDir != "" {
		candidates = append(candidates, filepath.Join(cacheDir, pkgName))
	}

	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		programData := os.Getenv("PROGRAMDATA")
		if localAppData != "" {
			candidates = append(candidates, filepath.Join(localAppData, pkgName))
		}
		if programData != "" {
			candidates = append(candidates, filepath.Join(programData, pkgName))
		}
	default: // linux, darwin
		candidates = append(candidates,
			filepath.Join(home, ".local", "share", pkgName),
			filepath.Join("/etc", pkgName),
			filepath.Join("/var", "lib", pkgName),
		)
	}

	var found []string
	for _, path := range candidates {
		if _, statErr := os.Stat(path); statErr == nil {
			found = append(found, path)
		}
	}
	return found
}

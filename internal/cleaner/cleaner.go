// Package cleaner orchestrates package uninstallation and leftover file detection.
package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/leftover"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Cleaner holds a map of detectors by manager name for fast lookup,
// along with the active whitelist used by leftover scanners.
type Cleaner struct {
	detectors map[string]detector.Detector
	whitelist *leftover.Whitelist
}

// New creates a Cleaner from a slice of detectors. The whitelist is
// loaded from the conventional config path; load errors fall back to an
// empty whitelist so a malformed file never blocks uninstall.
func New(detectors []detector.Detector) *Cleaner {
	m := make(map[string]detector.Detector, len(detectors))
	for _, d := range detectors {
		m[d.Name()] = d
	}
	wl, err := leftover.LoadWhitelist(leftover.DefaultWhitelistPath())
	if err != nil {
		wl = &leftover.Whitelist{}
	}
	return &Cleaner{detectors: m, whitelist: wl}
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
				attachLeftovers(&result, p, c.whitelist)
			}
		}

		results = append(results, result)
	}
	return results
}

// attachLeftovers runs the per-manager scanner and copies its findings
// into the result. Results are also flattened to the legacy
// Leftovers []string slice so callers that have not migrated yet keep
// working.
func attachLeftovers(result *pkg.UninstallResult, p pkg.Package, w *leftover.Whitelist) {
	scan := leftover.ScannerFor(p.Manager, w).Scan(p)
	result.LeftoverEntries = make([]pkg.LeftoverEntry, 0, len(scan.Entries))
	for _, e := range scan.Entries {
		result.LeftoverEntries = append(result.LeftoverEntries, pkg.LeftoverEntry{
			Path:        e.Path,
			Size:        e.Size,
			Whitelisted: e.Whitelisted,
		})
		result.Leftovers = append(result.Leftovers, e.Path)
	}
	result.LeftoverTotal = scan.Total
}

// FindLeftovers returns paths of leftover config/cache directories for
// a given package name, using platform-appropriate locations.
//
// Deprecated: prefer leftover.ScannerFor(manager) which returns rich
// metadata (size, whitelist status). Kept for any remaining callers.
func FindLeftovers(pkgName string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	configDir, _ := os.UserConfigDir()
	cacheDir, _ := os.UserCacheDir()

	candidates := []string{
		filepath.Join(home, "."+pkgName),
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
	default:
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

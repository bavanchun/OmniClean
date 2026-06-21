// Package cleanup aggregates classified packages across detectors into a list
// of removable cleanup candidates. It is distinct from internal/cleaner, which
// executes uninstalls; this package only reads and ranks.
package cleanup

import (
	"context"
	"sort"
	"time"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// DefaultDetectorTimeout bounds each detector's list+classify work so one slow
// or hung manager query degrades to "skipped" instead of freezing aggregation.
const DefaultDetectorTimeout = 8 * time.Second

// Aggregate lists and classifies packages across all given detectors and
// returns only the removable candidates (RoleManual / RoleOrphan), sorted with
// Orphan first, then oldest InstalledAt first (unknown install time last).
// Dependency and Unknown packages are never returned.
func Aggregate(ctx context.Context, detectors []detector.Detector) []pkg.Package {
	return aggregate(ctx, detectors, DefaultDetectorTimeout)
}

// aggregate is the timeout-injectable core, exercised directly by tests.
func aggregate(ctx context.Context, detectors []detector.Detector, timeout time.Duration) []pkg.Package {
	var candidates []pkg.Package
	for _, d := range detectors {
		candidates = append(candidates, collectOne(ctx, d, timeout)...)
	}
	sortCandidates(candidates)
	return candidates
}

// collectOne runs one detector's list+classify under a bounded timeout and
// returns its removable candidates, or nil on any error/timeout.
func collectOne(ctx context.Context, d detector.Detector, timeout time.Duration) []pkg.Package {
	dctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pkgs, err := d.ListPackages(dctx)
	if err != nil {
		return nil
	}
	classified, err := detector.ClassifyIfSupported(dctx, d, pkgs)
	if err != nil {
		return nil
	}

	var out []pkg.Package
	for _, p := range classified {
		if p.Role == pkg.RoleManual || p.Role == pkg.RoleOrphan {
			out = append(out, p)
		}
	}
	return out
}

// sortCandidates orders candidates Orphan-first, then by oldest InstalledAt
// (unknown install time sorts last within a role group).
func sortCandidates(c []pkg.Package) {
	sort.SliceStable(c, func(i, j int) bool {
		if pi, pj := rolePriority(c[i].Role), rolePriority(c[j].Role); pi != pj {
			return pi < pj
		}
		ti, tj := c[i].InstalledAt, c[j].InstalledAt
		izero, jzero := ti.IsZero(), tj.IsZero()
		if izero != jzero {
			return !izero // known install time before unknown
		}
		if izero { // both unknown: keep stable order
			return false
		}
		return ti.Before(tj) // oldest first
	})
}

// rolePriority ranks Orphan (safest to remove) ahead of Manual.
func rolePriority(r pkg.Role) int {
	if r == pkg.RoleOrphan {
		return 0
	}
	return 1
}

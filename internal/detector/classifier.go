package detector

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// classifyTimeout bounds each detector's Classify so a slow or hung manager
// query degrades to RoleUnknown instead of freezing the caller.
const classifyTimeout = 5 * time.Second

// statFunc resolves a path to its modification time. It is a seam so tests can
// inject install-time fixtures without touching the real filesystem.
type statFunc func(path string) (time.Time, error)

// osStatMtime is the production statFunc: it returns a path's mtime.
func osStatMtime(path string) (time.Time, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// lineSet parses newline-separated command output into a set of trimmed,
// non-empty tokens (the first whitespace-delimited field of each line).
func lineSet(out string) map[string]bool {
	set := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		set[fields[0]] = true
	}
	return set
}

// markAllUnknown returns a copy of pkgs with every Role reset to RoleUnknown.
// Used when a manager query is unavailable: safe-by-default, never removable.
func markAllUnknown(pkgs []pkg.Package) []pkg.Package {
	out := make([]pkg.Package, len(pkgs))
	copy(out, pkgs)
	for i := range out {
		out[i].Role = pkg.RoleUnknown
	}
	return out
}

// RemovableClassifier is an optional capability. Detectors that can consult
// their manager's dependency bookkeeping implement it; callers use a type
// assertion (d.(RemovableClassifier)) and skip detectors that don't.
type RemovableClassifier interface {
	// Classify annotates packages with Role and, when cheaply available,
	// InstalledAt. It is read-only: no uninstall side effects.
	Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error)
}

// ClassifyIfSupported runs the detector's classifier when it implements
// RemovableClassifier; otherwise it returns the packages unchanged (each
// keeps its zero-value RoleUnknown). This lets callers treat every detector
// uniformly without caring whether it can classify.
func ClassifyIfSupported(ctx context.Context, d Detector, pkgs []pkg.Package) ([]pkg.Package, error) {
	c, ok := d.(RemovableClassifier)
	if !ok {
		return pkgs, nil
	}
	return c.Classify(ctx, pkgs)
}

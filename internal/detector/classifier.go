package detector

import (
	"context"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

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

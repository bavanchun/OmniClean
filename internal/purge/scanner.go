package purge

import "context"

// Options tune Scan behaviour. Zero values use sensible defaults.
type Options struct {
	// MaxDepth bounds how deep we descend. 0 means unlimited.
	MaxDepth int
	// RecentDays marks targets newer than this as Recent. Defaults to 7.
	RecentDays int
	// IncludeStacks restricts the scan to specific stacks. Empty means
	// every stack in DefaultPatterns is scanned.
	IncludeStacks []Stack
}

// Scanner is the public scan entry point. Implementations may use fd
// (when available) or filepath.WalkDir; both yield the same Target slice.
type Scanner interface {
	Scan(ctx context.Context, roots []string, opts Options) ([]Target, error)
}

package analyze

import "context"

// Scanner walks a path and returns a Result. Implementations can vary
// the concurrency strategy; the TUI does not care.
type Scanner interface {
	Scan(ctx context.Context, path string, opts Options) (Result, error)
}

// Options tune the scanner. Zero values are sensible defaults.
type Options struct {
	// LargeFileTopN limits how many entries appear in Result.LargeFiles.
	// Defaults to 20 when 0.
	LargeFileTopN int
	// LargeFileMinBytes filters out files smaller than this from
	// LargeFiles. Defaults to 100 MiB when 0.
	LargeFileMinBytes int64
	// MaxWorkers caps the concurrent workers used to size subdirs.
	// Defaults to runtime.NumCPU() when 0.
	MaxWorkers int
}

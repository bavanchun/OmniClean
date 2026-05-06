package leftover

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// SizeLimits caps how much work pathSize will do for a single path. The
// zero value applies sensible defaults (500k entries, 10 MiB-cumulative
// stat budget). Callers can shrink the limits for very tight budgets.
type SizeLimits struct {
	MaxEntries int64
	MaxBytes   int64
}

// errBudget is sentinel returned when a walk hits SizeLimits and bails
// out early. It is intentionally unexported.
var errBudget = errors.New("leftover: size budget exceeded")

// pathSize returns the total size in bytes of the file or directory at
// path. Symlinks are not followed. If limits.MaxEntries is reached the
// partial total is returned along with a true "truncated" flag so the
// caller can mark the result as approximate.
func pathSize(path string, limits SizeLimits) (size int64, truncated bool) {
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 500_000
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 40 // 1 TiB — effectively unlimited
	}

	info, err := os.Lstat(path)
	if err != nil {
		return 0, false
	}
	if !info.IsDir() {
		return info.Size(), false
	}

	var total int64
	var entries int64
	walkErr := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip unreadable subtree, keep counting the rest.
			return nil
		}
		entries++
		if entries > limits.MaxEntries || total > limits.MaxBytes {
			truncated = true
			return errBudget
		}
		if d.IsDir() {
			return nil
		}
		fi, ferr := d.Info()
		if ferr != nil {
			return nil
		}
		total += fi.Size()
		return nil
	})
	if walkErr != nil && !errors.Is(walkErr, errBudget) {
		return total, truncated
	}
	return total, truncated
}

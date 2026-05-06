// Package leftover discovers files and directories that a package manager
// leaves behind after uninstalling a package. Each supported manager has
// its own Scanner implementation; the registry returns the right Scanner
// for a given pkg.ManagerType.
//
// Scanners are deliberately read-only. They never delete anything; the
// cleaner package decides what to do with the results.
package leftover

import (
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Entry is a single leftover path along with its computed size in bytes
// and whether the user's whitelist protects it from removal suggestions.
type Entry struct {
	Path        string
	Size        int64
	Whitelisted bool
}

// Result aggregates a scan run for one package.
type Result struct {
	Manager pkg.ManagerType
	Package string
	Entries []Entry
	Total   int64 // sum of Entries[].Size
	Skipped []string
}

// Scanner inspects the filesystem for leftover artifacts of a single
// package. Implementations must be safe to call concurrently.
type Scanner interface {
	// Manager returns the pkg.ManagerType this scanner targets.
	Manager() pkg.ManagerType
	// Scan walks known leftover locations for the given package name and
	// returns a Result. Errors at the per-path level are recorded in
	// Result.Skipped instead of aborting the whole scan.
	Scan(p pkg.Package) Result
}

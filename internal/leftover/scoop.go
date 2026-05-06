package leftover

import (
	"os"
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Scoop installs into %USERPROFILE%\scoop by default with caches under
// scoop\cache and per-app data under scoop\persist. SCOOP env var
// overrides the install root.
type Scoop struct{ Whitelist *Whitelist }

func (Scoop) Manager() pkg.ManagerType { return pkg.ManagerScoop }

func (s Scoop) Scan(p pkg.Package) Result {
	root := os.Getenv("SCOOP")
	if root == "" {
		if h, err := os.UserHomeDir(); err == nil {
			root = filepath.Join(h, "scoop")
		}
	}
	candidates := []string{}
	if root != "" {
		candidates = append(candidates,
			filepath.Join(root, "cache", p.Name),
			filepath.Join(root, "persist", p.Name),
			filepath.Join(root, "apps", p.Name),
		)
	}
	return scanCandidates(pkg.ManagerScoop, p.Name, candidates, s.Whitelist)
}

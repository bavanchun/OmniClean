package leftover

import (
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Snap stores per-user data under ~/snap/<name> and system data under
// /var/snap/<name>. The user-side directory is what end users typically
// want surfaced when uninstalling.
type Snap struct{ Whitelist *Whitelist }

func (Snap) Manager() pkg.ManagerType { return pkg.ManagerSnap }

func (s Snap) Scan(p pkg.Package) Result {
	up := resolveUserPaths()
	candidates := []string{}
	if up.Home != "" {
		candidates = append(candidates, filepath.Join(up.Home, "snap", p.Name))
	}
	candidates = append(candidates,
		filepath.Join("/var", "snap", p.Name),
	)
	return scanCandidates(pkg.ManagerSnap, p.Name, candidates, s.Whitelist)
}

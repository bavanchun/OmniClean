package leftover

import (
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// APT keeps system-owned residue under /var/cache/apt and /etc.
// We deliberately probe only paths a non-root user can stat to avoid
// surfacing entries the user cannot act on. The cleaner package decides
// whether to display them.
type APT struct{ Whitelist *Whitelist }

func (APT) Manager() pkg.ManagerType { return pkg.ManagerAPT }

func (a APT) Scan(p pkg.Package) Result {
	candidates := []string{
		filepath.Join("/var", "lib", "dpkg", "info", p.Name+".list"),
		filepath.Join("/var", "lib", "dpkg", "info", p.Name+".md5sums"),
		filepath.Join("/etc", p.Name),
	}
	up := resolveUserPaths()
	if up.Home != "" {
		candidates = append(candidates,
			filepath.Join(up.Home, "."+p.Name),
			filepath.Join(up.Home, ".config", p.Name),
		)
	}
	if up.Config != "" {
		candidates = append(candidates,
			filepath.Join(up.Config, p.Name),
		)
	}
	return scanCandidates(pkg.ManagerAPT, p.Name, candidates, a.Whitelist)
}

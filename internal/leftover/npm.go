package leftover

import (
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// NPM probes npm cache and global module residue.
//
//   - ~/.npm/_cacache/content-v2 holds tarball blobs but is hashed; we
//     instead point at ~/.npm/_logs/<name>* and the per-package folder
//     under ~/.npm/_cacache/index-v5 only when present at the top level.
//   - The global prefix dir (lib/node_modules/<name>) typically belongs
//     to root on POSIX so we just probe the user's npm prefix when set.
type NPM struct{ Whitelist *Whitelist }

func (NPM) Manager() pkg.ManagerType { return pkg.ManagerNPM }

func (n NPM) Scan(p pkg.Package) Result {
	up := resolveUserPaths()
	candidates := []string{}
	if up.Home != "" {
		candidates = append(candidates,
			filepath.Join(up.Home, ".npm", "_logs", p.Name),
			filepath.Join(up.Home, ".npm", p.Name),
			filepath.Join(up.Home, ".config", "configstore", p.Name+".json"),
			filepath.Join(up.Home, ".npm-packages", "lib", "node_modules", p.Name),
		)
	}
	if up.Cache != "" {
		candidates = append(candidates,
			filepath.Join(up.Cache, "npm", p.Name),
		)
	}
	return scanCandidates(pkg.ManagerNPM, p.Name, candidates, n.Whitelist)
}

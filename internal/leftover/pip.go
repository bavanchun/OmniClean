package leftover

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Pip probes pip's user-level wheel cache and per-user site-packages
// residue. The wheel cache is keyed by package name under
// ~/.cache/pip/wheels (Linux) or ~/Library/Caches/pip/wheels (macOS).
type Pip struct{ Whitelist *Whitelist }

func (Pip) Manager() pkg.ManagerType { return pkg.ManagerPip }

func (p Pip) Scan(pkgInfo pkg.Package) Result {
	up := resolveUserPaths()
	candidates := []string{}
	name := strings.ToLower(pkgInfo.Name)

	if up.Cache != "" {
		candidates = append(candidates,
			filepath.Join(up.Cache, "pip", "wheels", name),
			filepath.Join(up.Cache, "pip", "http", name),
		)
		if runtime.GOOS == "darwin" {
			candidates = append(candidates,
				filepath.Join(up.Cache, "pip", "wheels", pkgInfo.Name),
			)
		}
	}
	if up.Home != "" {
		// Probe the XDG cache path explicitly so macOS users with a
		// pip-managed Linux-style ~/.cache/pip layout (e.g. via WSL or
		// container mounts) still get a hit.
		candidates = append(candidates,
			filepath.Join(up.Home, ".cache", "pip", "wheels", name),
			filepath.Join(up.Home, ".cache", "pip", "http", name),
		)
		// User-site dist-info residue: name-VERSION.dist-info dirs are
		// matched by their parent; we only probe the legacy egg-info.
		candidates = append(candidates,
			filepath.Join(up.Home, ".local", "lib", "python3", "site-packages", name),
			filepath.Join(up.Home, ".local", "share", "pipx", "venvs", name),
		)
	}
	return scanCandidates(pkg.ManagerPip, pkgInfo.Name, candidates, p.Whitelist)
}

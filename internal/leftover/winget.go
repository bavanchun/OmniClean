package leftover

import (
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Winget surfaces per-user app data under %LOCALAPPDATA%\Packages\<id>.
// pkg.Name is treated as the package identifier reported by `winget
// list`, which usually matches the AppX package family name closely.
type Winget struct{ Whitelist *Whitelist }

func (Winget) Manager() pkg.ManagerType { return pkg.ManagerWinget }

func (w Winget) Scan(p pkg.Package) Result {
	up := resolveUserPaths()
	candidates := []string{}
	if up.LocalAppData != "" {
		candidates = append(candidates,
			filepath.Join(up.LocalAppData, "Packages", p.Name),
			filepath.Join(up.LocalAppData, "Programs", p.Name),
		)
	}
	if up.AppData != "" {
		candidates = append(candidates,
			filepath.Join(up.AppData, p.Name),
		)
	}
	return scanCandidates(pkg.ManagerWinget, p.Name, candidates, w.Whitelist)
}

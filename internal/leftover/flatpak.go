package leftover

import (
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Flatpak stores per-app user data under ~/.var/app/<id>. The pkg.Name
// is treated as the application id (e.g. org.gimp.GIMP), matching what
// the flatpak detector reports.
type Flatpak struct{ Whitelist *Whitelist }

func (Flatpak) Manager() pkg.ManagerType { return pkg.ManagerFlatpak }

func (f Flatpak) Scan(p pkg.Package) Result {
	up := resolveUserPaths()
	candidates := []string{}
	if up.Home != "" {
		candidates = append(candidates,
			filepath.Join(up.Home, ".var", "app", p.Name),
		)
	}
	if up.Data != "" {
		candidates = append(candidates,
			filepath.Join(up.Data, "flatpak", "app", p.Name),
		)
	}
	return scanCandidates(pkg.ManagerFlatpak, p.Name, candidates, f.Whitelist)
}

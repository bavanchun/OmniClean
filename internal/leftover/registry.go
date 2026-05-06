package leftover

import "github.com/bavanchun/OmniClean/internal/pkg"

// ScannerFor returns the per-manager Scanner for the given manager,
// configured with the supplied whitelist. Unknown managers fall back to
// a no-op Scanner that always returns an empty Result so callers do not
// need to special-case them.
func ScannerFor(mgr pkg.ManagerType, w *Whitelist) Scanner {
	switch mgr {
	case pkg.ManagerBrew:
		return Brew{Whitelist: w}
	case pkg.ManagerNPM:
		return NPM{Whitelist: w}
	case pkg.ManagerPip:
		return Pip{Whitelist: w}
	case pkg.ManagerCargo:
		return Cargo{Whitelist: w}
	case pkg.ManagerAPT:
		return APT{Whitelist: w}
	case pkg.ManagerSnap:
		return Snap{Whitelist: w}
	case pkg.ManagerFlatpak:
		return Flatpak{Whitelist: w}
	case pkg.ManagerChoco:
		return Choco{Whitelist: w}
	case pkg.ManagerScoop:
		return Scoop{Whitelist: w}
	case pkg.ManagerWinget:
		return Winget{Whitelist: w}
	}
	return noopScanner{mgr: mgr}
}

type noopScanner struct{ mgr pkg.ManagerType }

func (n noopScanner) Manager() pkg.ManagerType  { return n.mgr }
func (n noopScanner) Scan(p pkg.Package) Result { return Result{Manager: n.mgr, Package: p.Name} }

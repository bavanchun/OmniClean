package leftover

import (
	"path/filepath"
	"runtime"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Brew probes Homebrew formula and cask leftovers. macOS keeps caches in
// ~/Library/Caches/Homebrew and per-app data in
// ~/Library/Application Support; Linuxbrew uses ~/.cache/Homebrew.
type Brew struct{ Whitelist *Whitelist }

func (Brew) Manager() pkg.ManagerType { return pkg.ManagerBrew }

func (b Brew) Scan(p pkg.Package) Result {
	up := resolveUserPaths()
	candidates := []string{}

	if up.Home != "" {
		candidates = append(candidates,
			filepath.Join(up.Home, ".cache", "Homebrew", "downloads", p.Name),
		)
	}
	if up.Cache != "" {
		candidates = append(candidates,
			filepath.Join(up.Cache, "Homebrew", "downloads", p.Name),
		)
	}
	if runtime.GOOS == "darwin" && up.Home != "" {
		candidates = append(candidates,
			filepath.Join(up.Home, "Library", "Caches", "Homebrew", "downloads", p.Name),
			filepath.Join(up.Home, "Library", "Application Support", p.Name),
			filepath.Join(up.Home, "Library", "Logs", p.Name),
			filepath.Join(up.Home, "Library", "Preferences", p.Name+".plist"),
		)
	}

	return scanCandidates(pkg.ManagerBrew, p.Name, candidates, b.Whitelist)
}

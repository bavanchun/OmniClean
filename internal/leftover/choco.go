package leftover

import (
	"os"
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Choco probes Chocolatey's lib-bad recovery folder and per-package log
// directory. ChocolateyInstall env var wins; otherwise we use the
// conventional %ProgramData%\chocolatey path.
type Choco struct{ Whitelist *Whitelist }

func (Choco) Manager() pkg.ManagerType { return pkg.ManagerChoco }

func (c Choco) Scan(p pkg.Package) Result {
	root := os.Getenv("ChocolateyInstall")
	up := resolveUserPaths()
	if root == "" && up.ProgramData != "" {
		root = filepath.Join(up.ProgramData, "chocolatey")
	}
	candidates := []string{}
	if root != "" {
		candidates = append(candidates,
			filepath.Join(root, "lib-bad", p.Name),
			filepath.Join(root, "lib", p.Name),
			filepath.Join(root, "logs", "chocolatey-"+p.Name+".log"),
		)
	}
	return scanCandidates(pkg.ManagerChoco, p.Name, candidates, c.Whitelist)
}

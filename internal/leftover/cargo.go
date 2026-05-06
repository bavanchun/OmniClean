package leftover

import (
	"os"
	"path/filepath"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Cargo probes the user's Cargo home for crate caches. CARGO_HOME wins
// when set, otherwise we fall back to ~/.cargo.
type Cargo struct{ Whitelist *Whitelist }

func (Cargo) Manager() pkg.ManagerType { return pkg.ManagerCargo }

func (c Cargo) Scan(p pkg.Package) Result {
	root := os.Getenv("CARGO_HOME")
	if root == "" {
		if h, err := os.UserHomeDir(); err == nil {
			root = filepath.Join(h, ".cargo")
		}
	}
	candidates := []string{}
	if root != "" {
		// Registry caches are sharded by registry index; we probe the
		// crate folders that actually live underneath. Stat-only via
		// existing() drops the ones that aren't present.
		candidates = append(candidates,
			filepath.Join(root, "registry", "cache", "github.com-1ecc6299db9ec823", p.Name+"-"+p.Version+".crate"),
			filepath.Join(root, "registry", "src", "github.com-1ecc6299db9ec823", p.Name+"-"+p.Version),
			filepath.Join(root, "bin", p.Name),
		)
	}
	return scanCandidates(pkg.ManagerCargo, p.Name, candidates, c.Whitelist)
}

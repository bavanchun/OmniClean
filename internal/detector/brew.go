package detector

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Brew detects packages installed via Homebrew.
type Brew struct {
	run  CommandRunner
	stat statFunc // resolves Cellar mtime for best-effort InstalledAt
}

// NewBrew creates a Brew detector with the given command runner.
func NewBrew(run CommandRunner) *Brew {
	return &Brew{run: run, stat: osStatMtime}
}

func (b *Brew) Name() string    { return "brew" }
func (b *Brew) NeedsSudo() bool { return false }
func (b *Brew) Available() bool { return LookPath("brew") }

func (b *Brew) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("brew uninstall %s", p.Name)
}

func (b *Brew) UninstallExecCmd(_ pkg.Package) *exec.Cmd { return nil }

func (b *Brew) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := b.run(ctx, "brew", "list", "--versions")
	if err != nil {
		return nil, fmt.Errorf("brew list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		p := pkg.Package{
			Name:    parts[0],
			Manager: pkg.ManagerBrew,
		}
		if len(parts) >= 2 {
			p.Version = parts[len(parts)-1]
		}
		packages = append(packages, p)
	}
	return packages, nil
}

// Classify marks each package Manual/Orphan/Dependency from Homebrew's own
// bookkeeping using only read-only commands. `brew leaves` lists top-level
// formulae (Manual); `brew autoremove -n` dry-runs the orphan set (Orphan);
// anything installed but neither is a still-required Dependency. Bounded by a
// context timeout; any failure of the primary query degrades to RoleUnknown.
func (b *Brew) Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error) {
	ctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()

	leavesOut, err := b.run(ctx, "brew", "leaves")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	manual := lineSet(leavesOut)

	// Orphans are best-effort: a failure here still leaves Manual/Dependency valid.
	orphan := map[string]bool{}
	if orphanOut, err := b.run(ctx, "brew", "autoremove", "-n"); err == nil {
		orphan = parseBrewAutoremove(orphanOut)
	}

	// Resolve the Cellar root once for best-effort InstalledAt (mtime).
	cellar, _ := b.run(ctx, "brew", "--cellar")
	cellar = strings.TrimSpace(cellar)

	out := make([]pkg.Package, len(pkgs))
	copy(out, pkgs)
	for i := range out {
		name := out[i].Name
		switch {
		case orphan[name]:
			out[i].Role = pkg.RoleOrphan
		case manual[name]:
			out[i].Role = pkg.RoleManual
		default:
			out[i].Role = pkg.RoleDependency
		}
		if cellar != "" && b.stat != nil {
			if t, err := b.stat(filepath.Join(cellar, name)); err == nil {
				out[i].InstalledAt = t
			}
		}
	}
	return out, nil
}

// parseBrewAutoremove extracts formula names from `brew autoremove -n` output,
// skipping the "==> Would remove …" header lines.
func parseBrewAutoremove(out string) map[string]bool {
	set := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "==>") || strings.Contains(line, "Would") {
			continue
		}
		for _, f := range strings.Fields(line) {
			set[f] = true
		}
	}
	return set
}

func (b *Brew) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := b.run(ctx, "brew", "uninstall", p.Name)
	if err != nil {
		return fmt.Errorf("brew uninstall %s: %w", p.Name, err)
	}
	return nil
}

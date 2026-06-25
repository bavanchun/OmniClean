package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Pacman detects packages installed via pacman on Arch Linux systems.
type Pacman struct {
	run  CommandRunner
	stat statFunc // resolves local db dir mtime for best-effort InstalledAt
}

// NewPacman creates a Pacman detector with the given command runner.
func NewPacman(run CommandRunner) *Pacman {
	return &Pacman{run: run, stat: osStatMtime}
}

func (p *Pacman) Name() string    { return "pacman" }
func (p *Pacman) NeedsSudo() bool { return true }
func (p *Pacman) Available() bool { return LookPath("pacman") }

func (p *Pacman) DryRunCommand(pkg pkg.Package) string {
	return fmt.Sprintf("sudo pacman -Rns --print %s", pkg.Name)
}

func (p *Pacman) UninstallExecCmd(pkg pkg.Package) *exec.Cmd {
	return NewSudoExecCmd("sudo", "pacman", "-Rns", "--noconfirm", pkg.Name)
}

// ListPackages queries pacman for all installed packages.
func (p *Pacman) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := p.run(ctx, "pacman", "-Q")
	if err != nil {
		return nil, fmt.Errorf("pacman list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		packages = append(packages, pkg.Package{
			Name:    fields[0],
			Version: fields[1],
			Manager: pkg.ManagerPacman,
		})
	}
	return packages, nil
}

// Classify marks each package Manual/Orphan/Dependency from pacman's bookkeeping
// using read-only queries (no sudo).
func (p *Pacman) Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error) {
	ctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()

	manualOut, err := p.run(ctx, "pacman", "-Qeq")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	manual := lineSet(manualOut)

	orphan := map[string]bool{}
	// pacman -Qtdq exits with non-zero (typically 1) if there are no orphans.
	// We treat any error here as "no orphans" (safe-by-default).
	if orphanOut, err := p.run(ctx, "pacman", "-Qtdq"); err == nil {
		orphan = lineSet(orphanOut)
	}

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

		if p.stat != nil {
			dir := fmt.Sprintf("/var/lib/pacman/local/%s-%s", name, out[i].Version)
			if t, err := p.stat(dir); err == nil {
				out[i].InstalledAt = t
			}
		}
	}
	return out, nil
}

// Uninstall is not used since NeedsSudo=true; interactive TUI uses UninstallExecCmd instead.
func (p *Pacman) Uninstall(_ context.Context, pkg pkg.Package) error {
	return fmt.Errorf("pacman: use UninstallExecCmd for interactive sudo uninstall of %s", pkg.Name)
}

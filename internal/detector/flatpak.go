package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Flatpak detects packages installed via Flatpak.
type Flatpak struct {
	run  CommandRunner
	stat statFunc // resolves install-dir mtime for best-effort InstalledAt
}

// NewFlatpak creates a Flatpak detector with the given command runner.
func NewFlatpak(run CommandRunner) *Flatpak {
	return &Flatpak{run: run, stat: osStatMtime}
}

func (f *Flatpak) Name() string    { return "flatpak" }
func (f *Flatpak) NeedsSudo() bool { return false }
func (f *Flatpak) Available() bool { return LookPath("flatpak") }

func (f *Flatpak) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("flatpak uninstall -y %s", p.Name)
}

func (f *Flatpak) UninstallExecCmd(_ pkg.Package) *exec.Cmd { return nil }

func (f *Flatpak) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := f.run(ctx, "flatpak", "list", "--app", "--columns=application,version,size")
	if err != nil {
		return nil, fmt.Errorf("flatpak list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		p := pkg.Package{
			Name:    parts[0],
			Manager: pkg.ManagerFlatpak,
		}
		if len(parts) >= 2 {
			p.Version = parts[1]
		}
		if len(parts) >= 3 {
			p.Size = pkg.ParseHumanSize(parts[2])
		}
		packages = append(packages, p)
	}
	return packages, nil
}

// Classify marks flatpak apps Manual. Flatpak is leaf-only: there is no
// documented read-only/dry-run form of `flatpak uninstall --unused`, so we
// never run a mutating command to detect orphans (trustworthy-or-silent).
// Every app reported by `flatpak list --app` is a top-level Manual install;
// packages absent from that set are left RoleUnknown. Bounded by a context
// timeout; a failed query degrades to RoleUnknown. InstalledAt is best-effort
// from the system install-dir mtime (tracks last write, not true install time).
func (f *Flatpak) Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error) {
	ctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()

	listOut, err := f.run(ctx, "flatpak", "list", "--app", "--columns=application")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	manual := lineSet(listOut)

	out := make([]pkg.Package, len(pkgs))
	copy(out, pkgs)
	for i := range out {
		name := out[i].Name
		if !manual[name] {
			continue // leave RoleUnknown; no orphan/dependency signal for flatpak
		}
		out[i].Role = pkg.RoleManual
		if f.stat != nil {
			if t, err := f.stat(fmt.Sprintf("/var/lib/flatpak/app/%s", name)); err == nil {
				out[i].InstalledAt = t
			}
		}
	}
	return out, nil
}

func (f *Flatpak) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := f.run(ctx, "flatpak", "uninstall", "-y", p.Name)
	if err != nil {
		return fmt.Errorf("flatpak uninstall %s: %w", p.Name, err)
	}
	return nil
}

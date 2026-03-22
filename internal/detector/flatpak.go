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
	run CommandRunner
}

// NewFlatpak creates a Flatpak detector with the given command runner.
func NewFlatpak(run CommandRunner) *Flatpak {
	return &Flatpak{run: run}
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

func (f *Flatpak) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := f.run(ctx, "flatpak", "uninstall", "-y", p.Name)
	if err != nil {
		return fmt.Errorf("flatpak uninstall %s: %w", p.Name, err)
	}
	return nil
}

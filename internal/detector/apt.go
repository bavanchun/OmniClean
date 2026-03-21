package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// APT detects packages installed via apt/dpkg on Debian/Ubuntu systems.
type APT struct {
	run CommandRunner
}

// NewAPT creates an APT detector with the given command runner.
func NewAPT(run CommandRunner) *APT {
	return &APT{run: run}
}

func (a *APT) Name() string    { return "apt" }
func (a *APT) NeedsSudo() bool { return true }
func (a *APT) Available() bool { return LookPath("dpkg-query") }

func (a *APT) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("sudo apt-get remove -y %s", p.Name)
}

func (a *APT) UninstallExecCmd(p pkg.Package) *exec.Cmd {
	return NewSudoExecCmd("sudo", "apt-get", "remove", "-y", p.Name)
}

// ListPackages queries dpkg for all installed packages.
func (a *APT) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	// Format: Name\tVersion\tInstalled-Size(kB)\tDescription
	out, err := a.run(ctx, "dpkg-query",
		"--show",
		"--showformat=${Package}\t${Version}\t${Installed-Size}\t${binary:Summary}\n",
	)
	if err != nil {
		return nil, fmt.Errorf("apt list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 2 {
			continue
		}

		p := pkg.Package{
			Name:    parts[0],
			Version: parts[1],
			Manager: pkg.ManagerAPT,
		}
		if len(parts) >= 3 {
			if kb, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64); err == nil {
				p.Size = kb * 1024
			}
		}
		if len(parts) >= 4 {
			p.Description = parts[3]
		}
		packages = append(packages, p)
	}
	return packages, nil
}

// Uninstall is not used for APT since NeedsSudo=true; the TUI calls
// UninstallExecCmd + tea.Exec instead.
func (a *APT) Uninstall(_ context.Context, p pkg.Package) error {
	return fmt.Errorf("apt: use UninstallExecCmd for interactive sudo uninstall of %s", p.Name)
}

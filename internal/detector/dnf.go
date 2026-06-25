package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// DNF detects packages installed via dnf/yum on RedHat/Fedora systems.
type DNF struct {
	run CommandRunner
}

// NewDNF creates a DNF detector with the given command runner.
func NewDNF(run CommandRunner) *DNF {
	return &DNF{run: run}
}

func (d *DNF) Name() string    { return "dnf" }
func (d *DNF) NeedsSudo() bool { return true }

func (d *DNF) Available() bool {
	return LookPath("dnf") || LookPath("yum")
}

func (d *DNF) resolveBinary() string {
	if !LookPath("dnf") && LookPath("yum") {
		return "yum"
	}
	return "dnf"
}

func (d *DNF) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("sudo %s remove -y %s", d.resolveBinary(), p.Name)
}

func (d *DNF) UninstallExecCmd(p pkg.Package) *exec.Cmd {
	return NewSudoExecCmd("sudo", d.resolveBinary(), "remove", "-y", p.Name)
}

// ListPackages queries rpm for all installed packages.
func (d *DNF) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := d.run(ctx, "rpm", "-qa", "--qf", "%{NAME}\t%{VERSION}-%{RELEASE}\t%{SIZE}\t%{INSTALLTIME}\t%{SUMMARY}\n")
	if err != nil {
		return nil, fmt.Errorf("dnf list packages: %w", err)
	}

	var packages []pkg.Package
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 2 {
			continue
		}

		p := pkg.Package{
			Name:    parts[0],
			Version: parts[1],
			Manager: pkg.ManagerType("dnf"),
		}

		if len(parts) >= 3 {
			if sz, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64); err == nil {
				p.Size = sz
			}
		}

		if len(parts) >= 4 {
			if sec, err := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64); err == nil && sec > 0 {
				p.InstalledAt = time.Unix(sec, 0)
			}
		}

		if len(parts) >= 5 {
			p.Description = parts[4]
		}

		packages = append(packages, p)
	}
	return packages, nil
}

// Classify marks each package Manual/Orphan/Dependency from dnf/yum's bookkeeping
// using read-only queries (no sudo). Bounded by a context timeout; a failed query
// degrades to RoleUnknown.
func (d *DNF) Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error) {
	ctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()

	bin := d.resolveBinary()

	manualOut, err := d.run(ctx, bin, "repoquery", "--userinstalled", "--qf", "%{name}\n")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	manual := lineSet(manualOut)

	orphanOut, err := d.run(ctx, bin, "repoquery", "--unneeded", "--qf", "%{name}\n")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	orphan := lineSet(orphanOut)

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
	}
	return out, nil
}

// Uninstall is not used for DNF since NeedsSudo=true; the TUI calls
// UninstallExecCmd + tea.Exec instead.
func (d *DNF) Uninstall(_ context.Context, p pkg.Package) error {
	return fmt.Errorf("%s: use UninstallExecCmd for interactive sudo uninstall of %s", d.resolveBinary(), p.Name)
}

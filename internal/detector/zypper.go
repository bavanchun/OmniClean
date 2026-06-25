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

// Zypper detects packages installed via zypper on openSUSE systems.
type Zypper struct {
	run CommandRunner
}

// NewZypper creates a Zypper detector with the given command runner.
func NewZypper(run CommandRunner) *Zypper {
	return &Zypper{run: run}
}

func (z *Zypper) Name() string    { return "zypper" }
func (z *Zypper) NeedsSudo() bool { return true }
func (z *Zypper) Available() bool { return LookPath("zypper") }

func (z *Zypper) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("sudo zypper --non-interactive remove --dry-run %s", p.Name)
}

func (z *Zypper) UninstallExecCmd(p pkg.Package) *exec.Cmd {
	return NewSudoExecCmd("sudo", "zypper", "--non-interactive", "remove", p.Name)
}

// ListPackages queries rpm for all installed packages.
func (z *Zypper) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := z.run(ctx, "rpm", "-qa", "--qf", "%{NAME}\t%{VERSION}-%{RELEASE}\t%{SIZE}\t%{INSTALLTIME}\t%{SUMMARY}\n")
	if err != nil {
		return nil, fmt.Errorf("zypper list packages: %w", err)
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
			Manager: pkg.ManagerZypper,
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

// Classify marks each package Manual/Orphan/Dependency from zypper's bookkeeping
// using read-only queries (no sudo).
func (z *Zypper) Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error) {
	ctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()

	manualOut, err := z.run(ctx, "zypper", "--quiet", "search", "-si")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	manual := parseZypperTable(manualOut, true)

	orphanOut, err := z.run(ctx, "zypper", "--quiet", "packages", "--unneeded")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	orphan := parseZypperTable(orphanOut, false)

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

// parseZypperTable extracts package names from zypper pipe-separated tables.
func parseZypperTable(out string, onlyManual bool) map[string]bool {
	set := map[string]bool{}
	lines := strings.Split(out, "\n")
	nameColIdx := -1
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "---") || strings.Contains(line, "---+---") {
			continue
		}
		parts := strings.Split(line, "|")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		if nameColIdx == -1 {
			for idx, part := range parts {
				if part == "Name" {
					nameColIdx = idx
					break
				}
			}
			continue
		}

		if nameColIdx >= 0 && nameColIdx < len(parts) {
			if onlyManual {
				status := parts[0]
				if status != "i+" {
					continue
				}
			}
			name := parts[nameColIdx]
			if name != "" {
				set[name] = true
			}
		}
	}
	return set
}

// Uninstall is not used since NeedsSudo=true; interactive TUI uses UninstallExecCmd instead.
func (z *Zypper) Uninstall(_ context.Context, p pkg.Package) error {
	return fmt.Errorf("zypper: use UninstallExecCmd for interactive sudo uninstall of %s", p.Name)
}

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
	run  CommandRunner
	stat statFunc // resolves dpkg .list mtime for best-effort InstalledAt
}

// NewAPT creates an APT detector with the given command runner.
func NewAPT(run CommandRunner) *APT {
	return &APT{run: run, stat: osStatMtime}
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
	// Format: Status\tName\tVersion\tInstalled-Size(kB)\tDescription
	// Status is the two-letter dpkg state, e.g. "ii" (installed), "rc" (removed, config files remain).
	// We include ${db:Status-Abbrev} so we can skip packages that are no longer installed.
	out, err := a.run(ctx, "dpkg-query",
		"--show",
		"--showformat=${db:Status-Abbrev}\t${Package}\t${Version}\t${Installed-Size}\t${binary:Summary}\n",
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
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 3 {
			continue
		}

		// Only include packages with "ii" status (installed, not flagged for removal).
		// Packages in "rc" (removed, config-files remain) or other states are skipped.
		status := strings.TrimSpace(parts[0])
		if !strings.HasPrefix(status, "ii") {
			continue
		}

		p := pkg.Package{
			Name:    parts[1],
			Version: parts[2],
			Manager: pkg.ManagerAPT,
		}
		if len(parts) >= 4 {
			if kb, err := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64); err == nil {
				p.Size = kb * 1024
			}
		}
		if len(parts) >= 5 {
			p.Description = parts[4]
		}
		packages = append(packages, p)
	}
	return packages, nil
}

// Classify marks each package Manual/Orphan/Dependency from apt's bookkeeping
// using only read-only queries (no sudo). `apt-mark showmanual` lists manually
// installed packages (Manual); `apt-get autoremove --dry-run` reports the
// orphan set via "Remv {pkg}" lines (Orphan); installed-but-neither is a
// still-required Dependency. Bounded by a context timeout; a failed primary
// query degrades to RoleUnknown. InstalledAt is best-effort from the dpkg
// .list file mtime (tracks last file-list write, not true install time).
func (a *APT) Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error) {
	ctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()

	manualOut, err := a.run(ctx, "apt-mark", "showmanual")
	if err != nil {
		return markAllUnknown(pkgs), nil
	}
	manual := lineSet(manualOut)

	// Orphans are best-effort; --dry-run keeps this read-only and sudo-free.
	orphan := map[string]bool{}
	if orphanOut, err := a.run(ctx, "apt-get", "autoremove", "--dry-run"); err == nil {
		orphan = parseAptRemv(orphanOut)
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
		if a.stat != nil {
			if t, err := a.stat(fmt.Sprintf("/var/lib/dpkg/info/%s.list", name)); err == nil {
				out[i].InstalledAt = t
			}
		}
	}
	return out, nil
}

// parseAptRemv extracts package names from `apt-get autoremove --dry-run`
// output. apt prints one "Remv {pkg} [version]" line per package it would
// remove; a parse miss yields no orphans (safe), never a crash.
func parseAptRemv(out string) map[string]bool {
	set := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Remv ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			set[fields[1]] = true
		}
	}
	return set
}

// Uninstall is not used for APT since NeedsSudo=true; the TUI calls
// UninstallExecCmd + tea.Exec instead.
func (a *APT) Uninstall(_ context.Context, p pkg.Package) error {
	return fmt.Errorf("apt: use UninstallExecCmd for interactive sudo uninstall of %s", p.Name)
}

package detector

import (
	"context"
	"fmt"
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

func (a *APT) Name() string { return "apt" }

func (a *APT) Available() bool {
	return LookPath("dpkg-query")
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
			// dpkg reports size in kB; convert to bytes
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

// Uninstall removes a package using apt-get. Requires sudo.
func (a *APT) Uninstall(ctx context.Context, p pkg.Package, dryRun bool) error {
	args := []string{"apt-get", "remove", "-y", p.Name}
	if dryRun {
		fmt.Printf("[dry-run] sudo %s\n", strings.Join(args, " "))
		return nil
	}
	_, err := a.run(ctx, "sudo", args...)
	if err != nil {
		return fmt.Errorf("apt uninstall %s: %w", p.Name, err)
	}
	return nil
}

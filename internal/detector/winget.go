package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Winget detects packages installed via Windows Package Manager (winget).
type Winget struct {
	run CommandRunner
}

// NewWinget creates a Winget detector with the given command runner.
func NewWinget(run CommandRunner) *Winget {
	return &Winget{run: run}
}

func (w *Winget) Name() string    { return "winget" }
func (w *Winget) NeedsSudo() bool { return false }
func (w *Winget) Available() bool { return LookPath("winget") }

func (w *Winget) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("winget uninstall --id %s", p.Name)
}

func (w *Winget) UninstallExecCmd(_ pkg.Package) *exec.Cmd { return nil }

func (w *Winget) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	// winget list outputs a table; skip header lines (---, Name, Source columns)
	out, err := w.run(ctx, "winget", "list", "--disable-interactivity")
	if err != nil {
		return nil, fmt.Errorf("winget list packages: %w", err)
	}

	var packages []pkg.Package
	lines := strings.Split(out, "\n")
	headerPassed := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "---") {
			headerPassed = true
			continue
		}
		if !headerPassed || line == "" {
			continue
		}
		// Fields: Name  Id  Version  Available  Source
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		packages = append(packages, pkg.Package{
			Name:    fields[1], // use Id as canonical name for uninstall
			Version: fields[2],
			Manager: pkg.ManagerWinget,
		})
	}
	return packages, nil
}

func (w *Winget) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := w.run(ctx, "winget", "uninstall", "--id", p.Name, "--silent")
	if err != nil {
		return fmt.Errorf("winget uninstall %s: %w", p.Name, err)
	}
	return nil
}

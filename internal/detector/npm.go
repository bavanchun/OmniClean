package detector

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// NPM detects globally installed npm packages.
type NPM struct {
	run CommandRunner
}

// NewNPM creates an NPM detector with the given command runner.
func NewNPM(run CommandRunner) *NPM {
	return &NPM{run: run}
}

func (n *NPM) Name() string    { return "npm" }
func (n *NPM) NeedsSudo() bool { return false }
func (n *NPM) Available() bool { return LookPath("npm") }

func (n *NPM) DryRunCommand(p pkg.Package) string {
	return fmt.Sprintf("npm uninstall --global %s", p.Name)
}

func (n *NPM) UninstallExecCmd(_ pkg.Package) *exec.Cmd { return nil }

func (n *NPM) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := n.run(ctx, "npm", "list", "--global", "--depth=0", "--json")
	if err != nil && out == "" {
		return nil, fmt.Errorf("npm list packages: %w", err)
	}
	// npm may return non-zero due to peer dep warnings but still emit valid JSON.

	var result struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if parseErr := json.Unmarshal([]byte(out), &result); parseErr != nil {
		if err != nil {
			return nil, fmt.Errorf("npm list packages: command failed: %w; parse failed: %v", err, parseErr)
		}
		return nil, fmt.Errorf("npm list parse: %w", parseErr)
	}

	var packages []pkg.Package
	for name, info := range result.Dependencies {
		packages = append(packages, pkg.Package{
			Name:    name,
			Version: info.Version,
			Manager: pkg.ManagerNPM,
		})
	}
	return packages, nil
}

func (n *NPM) Uninstall(ctx context.Context, p pkg.Package) error {
	_, err := n.run(ctx, "npm", "uninstall", "--global", p.Name)
	if err != nil {
		return fmt.Errorf("npm uninstall %s: %w", p.Name, err)
	}
	return nil
}

package detector

import (
	"context"
	"encoding/json"
	"fmt"

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

func (n *NPM) Name() string { return "npm" }

func (n *NPM) Available() bool {
	return LookPath("npm")
}

func (n *NPM) ListPackages(ctx context.Context) ([]pkg.Package, error) {
	out, err := n.run(ctx, "npm", "list", "--global", "--depth=0", "--json")
	if err != nil {
		// npm list returns non-zero exit if there are peer dep issues; try to parse anyway
		if out == "" {
			return nil, fmt.Errorf("npm list packages: %w", err)
		}
	}

	var result struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil, fmt.Errorf("npm list parse: %w", err)
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

func (n *NPM) Uninstall(ctx context.Context, p pkg.Package, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] npm uninstall --global %s\n", p.Name)
		return nil
	}
	_, err := n.run(ctx, "npm", "uninstall", "--global", p.Name)
	if err != nil {
		return fmt.Errorf("npm uninstall %s: %w", p.Name, err)
	}
	return nil
}

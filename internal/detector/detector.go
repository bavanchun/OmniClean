// Package detector defines the Detector interface and shared utilities
// for all package manager implementations.
package detector

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// Detector is implemented by each supported package manager.
// Available() is checked at startup; only available detectors are used.
type Detector interface {
	// Name returns the human-readable name of the package manager.
	Name() string
	// Available returns true if the package manager is installed on the system.
	Available() bool
	// ListPackages returns all packages installed by this manager.
	ListPackages(ctx context.Context) ([]pkg.Package, error)
	// Uninstall removes the given package. If dryRun is true, it prints the
	// command that would be run without executing it.
	Uninstall(ctx context.Context, p pkg.Package, dryRun bool) error
}

// CommandRunner executes an external command and returns its trimmed stdout.
// It is defined as a type so tests can inject a fake runner.
type CommandRunner func(ctx context.Context, name string, args ...string) (string, error)

// DefaultRunner is the CommandRunner used in production.
func DefaultRunner(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w\nstderr: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// LookPath wraps exec.LookPath so detectors can check binary availability.
func LookPath(file string) bool {
	_, err := exec.LookPath(file)
	return err == nil
}

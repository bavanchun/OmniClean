// Package detector defines the Detector interface and shared utilities
// for all package manager implementations.
package detector

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
	// NeedsSudo returns true if uninstalling requires elevated privileges.
	NeedsSudo() bool
	// DryRunCommand returns the shell command string that would be executed,
	// without actually running it.
	DryRunCommand(p pkg.Package) string
	// UninstallExecCmd returns the *exec.Cmd to run for interactive uninstall
	// (e.g. with sudo password prompt). Only meaningful when NeedsSudo() is true.
	UninstallExecCmd(p pkg.Package) *exec.Cmd
	// ListPackages returns all packages installed by this manager.
	ListPackages(ctx context.Context) ([]pkg.Package, error)
	// Uninstall removes the given package. For managers requiring sudo,
	// the TUI handles execution via UninstallExecCmd instead.
	Uninstall(ctx context.Context, p pkg.Package) error
}

// CommandRunner executes an external command and returns its trimmed stdout.
// It is defined as a type so tests can inject a fake runner.
type CommandRunner func(ctx context.Context, name string, args ...string) (string, error)

// DefaultRunner is the CommandRunner used in production.
// It captures stdout/stderr into buffers (no stdin — cannot handle interactive prompts).
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

// NewSudoExecCmd builds an *exec.Cmd for an interactive sudo command.
// The returned command connects stdin/stdout/stderr to the real terminal,
// allowing sudo to prompt for a password.
func NewSudoExecCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// LookPath wraps exec.LookPath so detectors can check binary availability.
func LookPath(file string) bool {
	_, err := exec.LookPath(file)
	return err == nil
}

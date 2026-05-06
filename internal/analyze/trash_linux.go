//go:build linux

package analyze

import (
	"errors"
	"fmt"
	"os/exec"
)

// MoveToTrash on Linux prefers gio (GNOME, freedesktop) and falls back
// to trash-put from the trash-cli package. When neither tool is
// available we surface a clear error so callers can decide whether to
// prompt the user for a hard delete.
func MoveToTrash(path string) error {
	if p, err := exec.LookPath("gio"); err == nil {
		out, err := exec.Command(p, "trash", path).CombinedOutput()
		if err == nil {
			return nil
		}
		return fmt.Errorf("gio trash %s: %w (%s)", path, err, string(out))
	}
	if p, err := exec.LookPath("trash-put"); err == nil {
		out, err := exec.Command(p, path).CombinedOutput()
		if err == nil {
			return nil
		}
		return fmt.Errorf("trash-put %s: %w (%s)", path, err, string(out))
	}
	return errors.New("no trash helper available; install 'gio' (GNOME) or 'trash-cli'")
}

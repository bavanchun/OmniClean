//go:build darwin

package analyze

import (
	"fmt"
	"os/exec"
)

// MoveToTrash on macOS asks Finder to move the path to the user's
// Trash. This preserves the "Put Back" arrow and respects iCloud Drive
// trash routing automatically.
func MoveToTrash(path string) error {
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file %q`, path)
	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript trash %s: %w (%s)", path, err, string(out))
	}
	return nil
}

//go:build windows

package analyze

import (
	"fmt"
	"os/exec"
)

// MoveToTrash on Windows uses PowerShell to invoke Microsoft.VisualBasic
// FileSystem helpers, which is the documented way to send files to the
// Recycle Bin without writing native shell-API bindings. The path is
// quoted into a single-quoted PowerShell literal to avoid escape
// surprises with spaces or apostrophes.
func MoveToTrash(path string) error {
	script := fmt.Sprintf(`Add-Type -AssemblyName Microsoft.VisualBasic; ` +
		`[Microsoft.VisualBasic.FileIO.FileSystem]::DeleteDirectory('%s', ` +
		`'OnlyErrorDialogs', 'SendToRecycleBin')`, escapePS(path))
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("powershell trash %s: %w (%s)", path, err, string(out))
	}
	return nil
}

// escapePS doubles single quotes so a path containing one survives the
// PowerShell string literal we embed it in.
func escapePS(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			out = append(out, '\'', '\'')
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

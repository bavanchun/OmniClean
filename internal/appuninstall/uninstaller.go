//go:build darwin

package appuninstall

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// DeleteBundle removes the .app bundle. Tries osascript Finder trash first,
// falls back to os.RemoveAll. Returns DeleteResult with nil Err on success.
func DeleteBundle(ctx context.Context, b Bundle, dryRun bool) DeleteResult {
	if dryRun {
		return DeleteResult{Path: b.Path}
	}
	// Try Finder trash via osascript
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file %q`, b.Path)
	if err := exec.CommandContext(ctx, "osascript", "-e", script).Run(); err == nil {
		return DeleteResult{Path: b.Path}
	}
	// Fallback: permanent delete
	return DeleteResult{Path: b.Path, Err: os.RemoveAll(b.Path)}
}

// DeleteLeftovers removes a list of leftover entries using os.RemoveAll.
func DeleteLeftovers(ctx context.Context, entries []LeftoverEntry, dryRun bool) []DeleteResult {
	results := make([]DeleteResult, 0, len(entries))
	for _, e := range entries {
		if ctx.Err() != nil {
			results = append(results, DeleteResult{Path: e.Path, Err: ctx.Err()})
			continue
		}
		if dryRun {
			results = append(results, DeleteResult{Path: e.Path})
			continue
		}
		results = append(results, DeleteResult{Path: e.Path, Err: os.RemoveAll(e.Path)})
	}
	return results
}

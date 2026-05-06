//go:build darwin

package appuninstall

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// ScanRoots returns the default scan roots: /Applications and ~/Applications (if exists).
func ScanRoots() []string {
	roots := []string{"/Applications"}
	if home, err := os.UserHomeDir(); err == nil {
		userApps := filepath.Join(home, "Applications")
		if _, err := os.Stat(userApps); err == nil {
			roots = append(roots, userApps)
		}
	}
	return roots
}

// Scan walks roots (depth=1) and returns all .app bundles.
func Scan(ctx context.Context, roots []string) ([]Bundle, error) {
	var bundles []Bundle
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue // skip unreadable roots
		}
		for _, e := range entries {
			if ctx.Err() != nil {
				return bundles, ctx.Err()
			}
			if !e.IsDir() || !strings.HasSuffix(e.Name(), ".app") {
				continue
			}
			path := filepath.Join(root, e.Name())
			bundles = append(bundles, ParseBundle(ctx, path))
		}
	}
	return bundles, nil
}

// dirSize recursively sums the byte size of all files under path.
func dirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

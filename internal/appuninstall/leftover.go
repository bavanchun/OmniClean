//go:build darwin

package appuninstall

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// FindLeftovers returns orphan paths for a given bundle.
// Returns empty slice (not error) when BundleID is empty.
func FindLeftovers(ctx context.Context, b Bundle) ([]LeftoverEntry, error) {
	if b.BundleID == "" {
		return nil, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	lib := filepath.Join(home, "Library")

	type probe struct {
		desc    string
		exact   string // exact path to check
		glob    string // glob pattern relative to a dir
		globDir string // dir to glob in
	}

	probes := []probe{
		{desc: "Preferences", globDir: filepath.Join(lib, "Preferences"), glob: "*" + b.BundleID + "*"},
		{desc: "Caches", exact: filepath.Join(lib, "Caches", b.BundleID)},
		{desc: "App Support", exact: filepath.Join(lib, "Application Support", b.Name)},
		{desc: "Saved State", exact: filepath.Join(lib, "Saved Application State", b.BundleID+".savedState")},
		{desc: "Cookies", globDir: filepath.Join(lib, "Cookies"), glob: "*" + b.BundleID + "*"},
		{desc: "WebKit", exact: filepath.Join(lib, "WebKit", b.BundleID)},
		{desc: "Containers", exact: filepath.Join(lib, "Containers", b.BundleID)},
		{desc: "Group Containers", globDir: filepath.Join(lib, "Group Containers"), glob: "*" + b.BundleID + "*"},
	}

	var entries []LeftoverEntry
	for _, p := range probes {
		if ctx.Err() != nil {
			break
		}
		if p.exact != "" {
			if info, err := os.Stat(p.exact); err == nil {
				size := info.Size()
				if info.IsDir() {
					size = dirSize(p.exact)
				}
				entries = append(entries, LeftoverEntry{Path: p.exact, Size: size, Desc: p.desc})
			}
		} else if p.glob != "" {
			matches, err := filepath.Glob(filepath.Join(p.globDir, p.glob))
			if err != nil || matches == nil {
				continue
			}
			for _, m := range matches {
				// skip if m contains the bundle app itself
				if strings.Contains(m, ".app") {
					continue
				}
				info, err := os.Stat(m)
				if err != nil {
					continue
				}
				size := info.Size()
				if info.IsDir() {
					size = dirSize(m)
				}
				entries = append(entries, LeftoverEntry{Path: m, Size: size, Desc: p.desc})
			}
		}
	}
	return entries, nil
}

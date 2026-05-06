package leftover

import (
	"os"
	"path/filepath"
	"runtime"
)

// userPaths holds the well-known per-user directories scanners commonly
// probe. Empty fields mean the OS could not provide a value, in which
// case the field should be skipped silently.
type userPaths struct {
	Home   string
	Config string
	Cache  string
	Data   string
	// Windows-specific roots; empty on other OSes.
	LocalAppData string
	AppData      string
	ProgramData  string
}

// resolveUserPaths fills in best-effort directory locations for the
// current user. It calls os.UserHomeDir, os.UserConfigDir and
// os.UserCacheDir which already honor XDG_* on Linux, the Library
// hierarchy on macOS, and APPDATA/LOCALAPPDATA on Windows.
func resolveUserPaths() userPaths {
	p := userPaths{}
	if h, err := os.UserHomeDir(); err == nil {
		p.Home = h
	}
	if c, err := os.UserConfigDir(); err == nil {
		p.Config = c
	}
	if c, err := os.UserCacheDir(); err == nil {
		p.Cache = c
	}
	if p.Home != "" {
		switch runtime.GOOS {
		case "linux":
			if v := os.Getenv("XDG_DATA_HOME"); v != "" {
				p.Data = v
			} else {
				p.Data = filepath.Join(p.Home, ".local", "share")
			}
		case "darwin":
			p.Data = filepath.Join(p.Home, "Library", "Application Support")
		}
	}
	p.LocalAppData = os.Getenv("LOCALAPPDATA")
	p.AppData = os.Getenv("APPDATA")
	p.ProgramData = os.Getenv("PROGRAMDATA")
	return p
}

// existing returns each candidate path that currently exists on disk.
// It silently drops empty strings and stat errors so callers can pass a
// noisy candidate list without pre-filtering.
func existing(candidates []string) []string {
	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, err := os.Stat(c); err == nil {
			out = append(out, c)
		}
	}
	return out
}

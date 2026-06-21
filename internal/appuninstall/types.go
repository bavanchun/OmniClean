//go:build darwin

package appuninstall

import "time"

// Bundle represents a discovered .app bundle
type Bundle struct {
	Path        string
	Name        string // display name
	BundleID    string // com.example.foo (empty if plist unreadable)
	Version     string // short version string
	Size        int64  // bytes
	LastModTime time.Time
}

// LeftoverEntry is one orphan path found after bundle removal
type LeftoverEntry struct {
	Path string
	Size int64
	Desc string // "Preferences", "Caches", "App Support", etc.
}

// DeleteResult captures outcome of one delete operation
type DeleteResult struct {
	Path string
	Err  error
}

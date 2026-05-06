// Package analyze implements the disk explorer that powers
// `omniclean analyze`. It walks a directory tree concurrently, totals
// per-entry sizes, surfaces the largest files, and emits both TUI and
// JSON-friendly views.
package analyze

import "time"

// DirEntry is a directory or file directly under the path being viewed.
// Size is the recursive byte total for directories.
type DirEntry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	IsDir      bool      `json:"is_dir"`
	LastAccess time.Time `json:"last_access,omitempty"`
}

// FileEntry is a single large file surfaced separately so users can act
// on individual blobs without expanding the directory it lives in.
type FileEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// Result is what Scan returns for a single directory analysis.
type Result struct {
	Path       string      `json:"path"`
	Entries    []DirEntry  `json:"entries"`
	LargeFiles []FileEntry `json:"large_files"`
	TotalSize  int64       `json:"total_size"`
	TotalFiles int64       `json:"total_files"`
}

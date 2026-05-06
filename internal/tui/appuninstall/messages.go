//go:build darwin

package appuninstall

import "github.com/bavanchun/OmniClean/internal/appuninstall"

// scanDoneMsg is delivered when the initial .app bundle scan finishes.
type scanDoneMsg struct {
	bundles []appuninstall.Bundle
	err     error
}

// bundleDeleteDoneMsg is delivered after all selected bundles have been
// processed (moved to Trash or permanently deleted).
type bundleDeleteDoneMsg struct {
	results []appuninstall.DeleteResult
}

// leftoverScanDoneMsg is delivered after leftover paths have been discovered
// for all deleted bundles.
type leftoverScanDoneMsg struct {
	entries []appuninstall.LeftoverEntry
	err     error
}

// leftoverDeleteDoneMsg is delivered after all leftover paths have been
// processed.
type leftoverDeleteDoneMsg struct {
	results []appuninstall.DeleteResult
}

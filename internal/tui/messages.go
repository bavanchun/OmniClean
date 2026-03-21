package tui

import "github.com/bavanchun/OmniClean/internal/pkg"

// PackagesLoadedMsg is sent when all detectors have finished listing packages.
type PackagesLoadedMsg struct {
	Packages []pkg.Package
}

// PackagesLoadErrorMsg is sent when package detection fails.
type PackagesLoadErrorMsg struct {
	Err error
}

// UninstallRequestMsg is sent when the user confirms packages to remove.
type UninstallRequestMsg struct {
	Packages []pkg.Package
	DryRun   bool
}

// UninstallCompleteMsg is sent when all uninstall operations have finished.
type UninstallCompleteMsg struct {
	Results []pkg.UninstallResult
}

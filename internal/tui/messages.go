package tui

import "github.com/bavanchun/OmniClean/internal/pkg"

// PackagesLoadedMsg is sent when all detectors have finished listing packages.
type PackagesLoadedMsg struct {
	Packages []pkg.Package
	Warnings []string // non-fatal errors from individual detectors
}

// PackagesLoadErrorMsg is sent when ALL detectors fail (no packages at all).
type PackagesLoadErrorMsg struct {
	Err error
}

// UninstallCompleteMsg is sent when all non-sudo uninstall operations finish.
type UninstallCompleteMsg struct {
	Results []pkg.UninstallResult
}

// sudoUninstallDoneMsg is sent after one sudo package is processed via tea.Exec.
type sudoUninstallDoneMsg struct {
	result pkg.UninstallResult
}

// DetectorDoneMsg is sent when a single detector finishes scanning.
type DetectorDoneMsg struct {
	Name     string
	Packages []pkg.Package
	Err      error
}

// allDetectorsDoneMsg is sent when the detector channel is closed (all done).
type allDetectorsDoneMsg struct{}

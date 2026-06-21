package cleanup

import "github.com/bavanchun/OmniClean/internal/pkg"

// loadDoneMsg is delivered when the async aggregation finishes.
type loadDoneMsg struct {
	candidates []pkg.Package
}

// normalDeleteDoneMsg is delivered after all non-sudo removals complete.
type normalDeleteDoneMsg struct {
	results []pkg.UninstallResult
}

// sudoDeleteDoneMsg is delivered after one sudo package is processed via tea.Exec.
type sudoDeleteDoneMsg struct {
	result pkg.UninstallResult
}

// sudoAllDoneMsg is delivered when the sudo queue is fully drained.
type sudoAllDoneMsg struct{}

package purge

import "github.com/bavanchun/OmniClean/internal/purge"

// scanDoneMsg is delivered when the scanner finishes (success or
// failure). Targets is empty when err is non-nil.
type scanDoneMsg struct {
	Targets []purge.Target
	Err     error
}

// deleteDoneMsg signals that a single target has been removed (or
// failed to remove). It is published per-target so the UI can update a
// running progress bar.
type deleteDoneMsg struct {
	Target purge.Target
	Err    error
}

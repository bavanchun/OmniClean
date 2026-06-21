package detector

import (
	"context"
	"strings"
	"time"
)

// fakeResponse pairs an args matcher with the output it should return.
type fakeResponse struct {
	match  func(name string, args []string) bool
	output string
	err    error
}

// fakeRunner is an args-switching, call-recording CommandRunner for classify
// tests. Unlike the single-canned-string fakes used by ListPackages tests, it
// dispatches different output per command (e.g. `brew leaves` vs
// `brew autoremove -n`) and records every invocation so a test can assert that
// only read-only commands were issued.
type fakeRunner struct {
	responses []fakeResponse
	calls     [][]string // each entry is name followed by its args
}

// run implements the CommandRunner signature.
func (f *fakeRunner) run(_ context.Context, name string, args ...string) (string, error) {
	call := append([]string{name}, args...)
	f.calls = append(f.calls, call)
	for _, r := range f.responses {
		if r.match(name, args) {
			return r.output, r.err
		}
	}
	return "", nil
}

// commandLine renders a recorded call as a single string for assertions.
func (f *fakeRunner) commandLine(i int) string {
	return strings.Join(f.calls[i], " ")
}

// containsArg reports whether a recorded call (name + args) contains exactly arg.
func containsArg(call []string, arg string) bool {
	for _, a := range call {
		if a == arg {
			return true
		}
	}
	return false
}

// argContains builds a matcher that fires when any arg contains sub.
func argContains(sub string) func(string, []string) bool {
	return func(_ string, args []string) bool {
		for _, a := range args {
			if strings.Contains(a, sub) {
				return true
			}
		}
		return false
	}
}

// fixedStat builds a statFunc that returns t for any path (install-time present).
func fixedStat(t time.Time) statFunc {
	return func(string) (time.Time, error) { return t, nil }
}

// errStat builds a statFunc that always fails (install-time absent).
func errStat() statFunc {
	return func(string) (time.Time, error) { return time.Time{}, context.DeadlineExceeded }
}

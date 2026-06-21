package detector

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

// TestCommandEnv_ForcesCLocale verifies the production runner pins the C locale
// so localized package-manager output (e.g. apt's "Remv") stays parseable,
// while preserving the rest of the inherited environment (PATH, etc.).
func TestCommandEnv_ForcesCLocale(t *testing.T) {
	env := commandEnv()

	has := func(want string) bool {
		for _, e := range env {
			if e == want {
				return true
			}
		}
		return false
	}
	if !has("LC_ALL=C") {
		t.Error("commandEnv() missing LC_ALL=C")
	}
	if !has("LANG=C") {
		t.Error("commandEnv() missing LANG=C")
	}

	// Inherited env must be preserved — assert a PATH entry survived.
	foundPath := false
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Error("commandEnv() dropped the inherited environment (no PATH=)")
	}
}

// TestDefaultRunner_RunsUnderCLocale proves the locale env actually reaches the
// child process. Unix-only (depends on `sh`); skipped on Windows.
func TestDefaultRunner_RunsUnderCLocale(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no POSIX sh on Windows")
	}
	out, err := DefaultRunner(context.Background(), "sh", "-c", "echo $LC_ALL")
	if err != nil {
		t.Fatalf("DefaultRunner error: %v", err)
	}
	if out != "C" {
		t.Errorf("child LC_ALL = %q, want C", out)
	}
}

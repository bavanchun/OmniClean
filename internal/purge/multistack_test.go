package purge

import (
	"context"
	"path/filepath"
	"testing"
)

// TestWalkerMatchesAllStacks seeds one artifact per stack and confirms
// every Pattern in DefaultPatterns surfaces correctly. Acts as a guard
// against regressions when new patterns are added.
func TestWalkerMatchesAllStacks(t *testing.T) {
	tmp := t.TempDir()

	// Seed one of each pattern.
	for _, p := range DefaultPatterns {
		writeFile(t, filepath.Join(tmp, "proj", p.Name, "marker"), 32)
	}

	got, err := NewWalker().Scan(context.Background(), []string{tmp}, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	seen := make(map[string]bool, len(got))
	for _, t := range got {
		seen[t.Pattern] = true
	}
	for _, p := range DefaultPatterns {
		if !seen[p.Name] {
			t.Errorf("expected pattern %s to be detected", p.Name)
		}
	}
}

func TestStackFilterRestricts(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "proj", "node_modules", "x.js"), 16)
	writeFile(t, filepath.Join(tmp, "proj", "target", "build.bin"), 16)

	got, err := NewWalker().Scan(context.Background(), []string{tmp}, Options{
		IncludeStacks: []Stack{StackRust},
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 1 || got[0].Stack != StackRust {
		t.Fatalf("expected only rust target, got %+v", got)
	}
}

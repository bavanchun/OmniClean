package purge

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path string, n int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, make([]byte, n), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestWalkerFindsNodeModules(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "proj-a", "node_modules", "react", "index.js"), 1024)
	writeFile(t, filepath.Join(tmp, "proj-a", "src", "main.ts"), 64)
	writeFile(t, filepath.Join(tmp, "proj-b", "node_modules", "lodash", "lodash.js"), 4096)

	got, err := NewWalker().Scan(context.Background(), []string{tmp}, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 node_modules targets, got %d: %+v", len(got), got)
	}
	for _, tgt := range got {
		if tgt.Pattern != "node_modules" || tgt.Stack != StackNode {
			t.Errorf("unexpected target: %+v", tgt)
		}
		if tgt.Size == 0 {
			t.Errorf("expected non-zero size for %s", tgt.Path)
		}
	}
}

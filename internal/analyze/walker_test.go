package analyze

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

func TestWalkerSizesImmediateChildren(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a", "x.bin"), 1024)
	writeFile(t, filepath.Join(root, "a", "y.bin"), 2048)
	writeFile(t, filepath.Join(root, "b.bin"), 4096)

	res, err := NewWalker().Scan(context.Background(), root, Options{
		LargeFileMinBytes: 1, // include everything for the test
		LargeFileTopN:     5,
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if res.TotalFiles != 3 {
		t.Fatalf("expected 3 files, got %d", res.TotalFiles)
	}
	if res.TotalSize != 1024+2048+4096 {
		t.Fatalf("unexpected total: %d", res.TotalSize)
	}
	if len(res.Entries) != 2 {
		t.Fatalf("expected 2 immediate children, got %d", len(res.Entries))
	}
	if res.Entries[0].Size < res.Entries[1].Size {
		t.Errorf("expected entries sorted desc by size; got %+v", res.Entries)
	}
}

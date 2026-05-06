package analyze

import (
	"context"
	"path/filepath"
	"testing"
)

// TestLargeFilesTopN seeds files of varying sizes and verifies the
// scanner returns at most TopN entries sorted descending, all above
// LargeFileMinBytes.
func TestLargeFilesTopN(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "tiny.bin"), 100)
	writeFile(t, filepath.Join(root, "med.bin"), 1024)
	writeFile(t, filepath.Join(root, "big1.bin"), 4096)
	writeFile(t, filepath.Join(root, "big2.bin"), 8192)
	writeFile(t, filepath.Join(root, "deep", "huge.bin"), 16384)

	res, err := NewWalker().Scan(context.Background(), root, Options{
		LargeFileTopN:     3,
		LargeFileMinBytes: 1024,
	})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(res.LargeFiles) != 3 {
		t.Fatalf("expected 3 large files, got %d: %+v", len(res.LargeFiles), res.LargeFiles)
	}
	for i := 0; i < len(res.LargeFiles)-1; i++ {
		if res.LargeFiles[i].Size < res.LargeFiles[i+1].Size {
			t.Errorf("expected desc order; got %+v", res.LargeFiles)
		}
	}
	if res.LargeFiles[0].Name != "huge.bin" {
		t.Errorf("expected biggest file first; got %s", res.LargeFiles[0].Name)
	}
}

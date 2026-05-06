//go:build darwin

package appuninstall

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanRoots(t *testing.T) {
	roots := ScanRoots()
	if len(roots) == 0 {
		t.Fatal("expected at least one root")
	}
	found := false
	for _, r := range roots {
		if r == "/Applications" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected /Applications in roots, got %v", roots)
	}
}

func TestScanEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	bundles, err := Scan(context.Background(), []string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 0 {
		t.Errorf("expected 0 bundles, got %d", len(bundles))
	}
}

func TestScanFindsApp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal .app directory structure
	appDir := filepath.Join(tmpDir, "Fake.app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}

	bundles, err := Scan(context.Background(), []string{tmpDir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	if bundles[0].Name != "Fake" {
		t.Errorf("expected Name=Fake, got %q", bundles[0].Name)
	}
	if bundles[0].Path != appDir {
		t.Errorf("expected Path=%q, got %q", appDir, bundles[0].Path)
	}
}

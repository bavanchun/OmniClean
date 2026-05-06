//go:build darwin

package appuninstall

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseBundleNoInfoPlist(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "MyApp.app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}

	b := ParseBundle(context.Background(), appDir)

	if b.Name != "MyApp" {
		t.Errorf("expected Name=MyApp, got %q", b.Name)
	}
	if b.BundleID != "" {
		t.Errorf("expected empty BundleID, got %q", b.BundleID)
	}
	if b.Version != "" {
		t.Errorf("expected empty Version, got %q", b.Version)
	}
	if b.Path != appDir {
		t.Errorf("expected Path=%q, got %q", appDir, b.Path)
	}
}

func TestParseBundleWithPlist(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "Test.app")
	contentsDir := filepath.Join(appDir, "Contents")
	if err := os.MkdirAll(contentsDir, 0o755); err != nil {
		t.Fatalf("failed to create Contents dir: %v", err)
	}

	// Write a minimal XML plist that plutil can parse
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>com.test.app</string>
<key>CFBundleShortVersionString</key><string>1.0</string>
<key>CFBundleDisplayName</key><string>Test App</string>
</dict></plist>`

	plistPath := filepath.Join(contentsDir, "Info.plist")
	if err := os.WriteFile(plistPath, []byte(plistContent), 0o644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	b := ParseBundle(context.Background(), appDir)

	if b.BundleID != "com.test.app" {
		t.Errorf("expected BundleID=com.test.app, got %q", b.BundleID)
	}
	if b.Version != "1.0" {
		t.Errorf("expected Version=1.0, got %q", b.Version)
	}
	if b.Name != "Test App" {
		t.Errorf("expected Name=Test App, got %q", b.Name)
	}
}

//go:build darwin

package appuninstall

import (
	"context"
	"testing"
)

func TestFindLeftoversNoBundleID(t *testing.T) {
	entries, err := FindLeftovers(context.Background(), Bundle{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil slice, got %v", entries)
	}
}

func TestFindLeftoversNoFiles(t *testing.T) {
	b := Bundle{
		Name:     "NonexistentXYZTest",
		BundleID: "com.nonexistent.xyz.test",
	}
	entries, err := FindLeftovers(context.Background(), b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 leftovers for nonexistent bundle, got %d: %v", len(entries), entries)
	}
}

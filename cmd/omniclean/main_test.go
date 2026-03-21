package main

import (
	"context"
	"os/exec"
	"testing"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
)

// stubDetector implements detector.Detector for testing filterDetectors.
type stubDetector struct{ name string }

func (s *stubDetector) Name() string                                          { return s.name }
func (s *stubDetector) Available() bool                                       { return true }
func (s *stubDetector) NeedsSudo() bool                                       { return false }
func (s *stubDetector) DryRunCommand(_ pkg.Package) string                    { return "" }
func (s *stubDetector) UninstallExecCmd(_ pkg.Package) *exec.Cmd              { return nil }
func (s *stubDetector) ListPackages(_ context.Context) ([]pkg.Package, error) { return nil, nil }
func (s *stubDetector) Uninstall(_ context.Context, _ pkg.Package) error      { return nil }

var _ detector.Detector = (*stubDetector)(nil)

func TestFilterDetectors(t *testing.T) {
	all := []detector.Detector{
		&stubDetector{"apt"},
		&stubDetector{"brew"},
		&stubDetector{"pip"},
	}

	tests := []struct {
		names   []string
		wantLen int
	}{
		{[]string{"apt"}, 1},
		{[]string{"apt", "pip"}, 2},
		{[]string{"brew"}, 1},
		{[]string{"unknown"}, 0},
		{[]string{}, 0},
	}

	for _, tc := range tests {
		got := filterDetectors(all, tc.names)
		if len(got) != tc.wantLen {
			t.Errorf("filterDetectors(%v) = %d detectors, want %d", tc.names, len(got), tc.wantLen)
		}
	}
}

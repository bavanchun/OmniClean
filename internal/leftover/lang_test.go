package leftover

import (
	"path/filepath"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestPipScanFindsWheelCache(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	makeFile(t, tmp, ".cache/pip/wheels/requests/wheel.whl", 2048)
	res := Pip{}.Scan(pkg.Package{Name: "requests", Manager: pkg.ManagerPip})
	if res.Total == 0 {
		t.Fatalf("expected non-zero pip leftover total; entries=%v", res.Entries)
	}
}

func TestCargoScanRespectsCargoHome(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	cargoHome := filepath.Join(tmp, "custom-cargo")
	t.Setenv("CARGO_HOME", cargoHome)
	makeFile(t, cargoHome, "bin/ripgrep", 4096)
	res := Cargo{}.Scan(pkg.Package{Name: "ripgrep", Manager: pkg.ManagerCargo, Version: "13.0.0"})
	if len(res.Entries) == 0 {
		t.Fatalf("expected cargo bin leftover; got %v", res.Entries)
	}
}

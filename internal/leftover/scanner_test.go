package leftover

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// makeFile creates a file with n bytes under dir/name, ensuring parents.
func makeFile(t *testing.T, dir, name string, n int) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, make([]byte, n), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

// withFakeHome points HOME (and XDG_*/UserCacheDir paths on Linux) at
// tmp so resolveUserPaths returns predictable values inside the test.
func withFakeHome(t *testing.T, tmp string) {
	t.Helper()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, ".local", "share"))
}

func TestBrewScanFindsHomebrewCache(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)

	want := makeFile(t, tmp, ".cache/Homebrew/downloads/wget/blob.tgz", 1024)
	res := Brew{}.Scan(pkg.Package{Name: "wget", Manager: pkg.ManagerBrew})

	if res.Total == 0 {
		t.Fatalf("expected non-zero total, got 0; entries=%v", res.Entries)
	}
	found := false
	for _, e := range res.Entries {
		if filepath.Dir(e.Path) == filepath.Dir(filepath.Dir(want)) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected to find brew downloads dir; got %v", res.Entries)
	}
}

func TestNPMScanFindsLogsAndConfigstore(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)

	makeFile(t, tmp, ".npm/_logs/lodash/2024-01-01.log", 512)
	makeFile(t, tmp, ".config/configstore/lodash.json", 64)
	res := NPM{}.Scan(pkg.Package{Name: "lodash", Manager: pkg.ManagerNPM})
	if len(res.Entries) < 2 {
		t.Fatalf("expected at least 2 leftover entries, got %d: %+v", len(res.Entries), res.Entries)
	}
}

func TestWhitelistProtectsEntry(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	makeFile(t, tmp, ".npm/_logs/foo/log.txt", 32)

	wlPath := filepath.Join(tmp, "wl")
	if runtime.GOOS == "windows" {
		// Use an absolute pattern that exists on disk so Match short-circuits.
		_ = os.WriteFile(wlPath, []byte(filepath.Join(tmp, ".npm", "_logs", "foo")+"\n"), 0o644)
	} else {
		_ = os.WriteFile(wlPath, []byte("~/.npm/_logs/foo\n"), 0o644)
	}
	w, err := LoadWhitelist(wlPath)
	if err != nil {
		t.Fatalf("LoadWhitelist: %v", err)
	}
	res := NPM{Whitelist: w}.Scan(pkg.Package{Name: "foo", Manager: pkg.ManagerNPM})
	if len(res.Entries) == 0 {
		t.Fatalf("expected at least one entry")
	}
	any := false
	for _, e := range res.Entries {
		if e.Whitelisted {
			any = true
		}
	}
	if !any {
		t.Errorf("expected an entry to be whitelisted; got %+v", res.Entries)
	}
}

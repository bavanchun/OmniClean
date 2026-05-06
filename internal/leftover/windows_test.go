package leftover

import (
	"path/filepath"
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestScoopScanRespectsSCOOPEnv(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	scoopRoot := filepath.Join(tmp, "scoop-custom")
	t.Setenv("SCOOP", scoopRoot)
	makeFile(t, scoopRoot, "cache/git/bundle.zip", 1024)
	res := Scoop{}.Scan(pkg.Package{Name: "git", Manager: pkg.ManagerScoop})
	if res.Total == 0 {
		t.Fatalf("expected scoop cache leftover; got %v", res.Entries)
	}
}

func TestChocoScanUsesChocolateyInstall(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	root := filepath.Join(tmp, "choco-install")
	t.Setenv("ChocolateyInstall", root)
	makeFile(t, root, "lib-bad/firefox/manifest.xml", 256)
	res := Choco{}.Scan(pkg.Package{Name: "firefox", Manager: pkg.ManagerChoco})
	if len(res.Entries) == 0 {
		t.Fatalf("expected choco lib-bad leftover; got %v", res.Entries)
	}
}

func TestWingetScanUsesLocalAppData(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "LocalAppData"))
	makeFile(t, filepath.Join(tmp, "LocalAppData"), "Packages/Microsoft.VSCode/state.json", 64)
	res := Winget{}.Scan(pkg.Package{Name: "Microsoft.VSCode", Manager: pkg.ManagerWinget})
	if len(res.Entries) == 0 {
		t.Fatalf("expected winget Packages leftover; got %v", res.Entries)
	}
}

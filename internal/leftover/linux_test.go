package leftover

import (
	"testing"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestSnapScanFindsUserDir(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	makeFile(t, tmp, "snap/firefox/common/state.json", 256)
	res := Snap{}.Scan(pkg.Package{Name: "firefox", Manager: pkg.ManagerSnap})
	if res.Total == 0 {
		t.Fatalf("expected snap leftover total > 0; got 0; entries=%v", res.Entries)
	}
}

func TestFlatpakScanFindsAppData(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	makeFile(t, tmp, ".var/app/org.gimp.GIMP/data/preferences", 128)
	res := Flatpak{}.Scan(pkg.Package{Name: "org.gimp.GIMP", Manager: pkg.ManagerFlatpak})
	if len(res.Entries) == 0 {
		t.Fatalf("expected flatpak app data entry; got %v", res.Entries)
	}
}

func TestAPTScanFindsConfigResidue(t *testing.T) {
	tmp := t.TempDir()
	withFakeHome(t, tmp)
	makeFile(t, tmp, ".config/htop/htoprc", 64)
	res := APT{}.Scan(pkg.Package{Name: "htop", Manager: pkg.ManagerAPT})
	if len(res.Entries) == 0 {
		t.Fatalf("expected apt user-config residue entry; got %v", res.Entries)
	}
}

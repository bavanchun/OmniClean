package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

func TestWriteCleanupJSON(t *testing.T) {
	installed := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)
	pkgs := []pkg.Package{
		{Name: "foo", Manager: pkg.ManagerBrew, Version: "1.2", Role: pkg.RoleOrphan, InstalledAt: installed, Size: 1048576},
		{Name: "bar", Manager: pkg.ManagerAPT, Version: "2.0", Role: pkg.RoleManual}, // zero InstalledAt + size
	}

	var buf bytes.Buffer
	if err := writeCleanupJSON(&buf, pkgs); err != nil {
		t.Fatalf("writeCleanupJSON error: %v", err)
	}
	out := buf.String()

	// Stable field names.
	for _, want := range []string{`"name"`, `"manager"`, `"version"`, `"role"`, `"size"`} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing field %s\n%s", want, out)
		}
	}
	// Role is lowercase from pkg.Role.
	if !strings.Contains(out, `"role": "orphan"`) {
		t.Errorf("expected lowercase role orphan\n%s", out)
	}
	if !strings.Contains(out, `"role": "manual"`) {
		t.Errorf("expected lowercase role manual\n%s", out)
	}
	// InstalledAt present (RFC3339, UTC) for foo.
	if !strings.Contains(out, `"installedAt": "2025-01-04T00:00:00Z"`) {
		t.Errorf("expected installedAt for foo\n%s", out)
	}

	// Decode and assert omitempty: bar (zero InstalledAt) must omit the field.
	var recs []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &recs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("got %d records, want 2", len(recs))
	}
	if _, ok := recs[1]["installedAt"]; ok {
		t.Errorf("bar has zero InstalledAt; field must be omitted, got %v", recs[1]["installedAt"])
	}
	if recs[0]["size"].(float64) != 1048576 {
		t.Errorf("foo size = %v, want 1048576", recs[0]["size"])
	}
	if recs[0]["manager"] != "brew" {
		t.Errorf("foo manager = %v, want brew", recs[0]["manager"])
	}
}

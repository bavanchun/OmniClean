package analyze

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestWriteJSONRoundTrip(t *testing.T) {
	in := Result{
		Path: "/tmp/foo",
		Entries: []DirEntry{
			{Name: "a", Path: "/tmp/foo/a", Size: 42, IsDir: true, LastAccess: time.Unix(1700000000, 0)},
		},
		LargeFiles: []FileEntry{
			{Name: "big.bin", Path: "/tmp/foo/big.bin", Size: 1 << 30},
		},
		TotalSize:  1 << 30,
		TotalFiles: 1,
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, in); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var out Result
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.TotalSize != in.TotalSize || out.TotalFiles != in.TotalFiles {
		t.Errorf("round-trip mismatch: %+v vs %+v", in, out)
	}
	if len(out.Entries) != 1 || out.Entries[0].Name != "a" {
		t.Errorf("entries lost in round-trip: %+v", out.Entries)
	}
}

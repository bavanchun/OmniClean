package pkg

import "testing"

func TestParseHumanSize(t *testing.T) {
	mb := int64(1024 * 1024)
	gb := int64(1024 * 1024 * 1024)
	tb := int64(1024 * 1024 * 1024 * 1024)

	tests := []struct {
		input string
		want  int64
	}{
		{"56.6 MB", int64(566 * mb / 10)},
		{"1.2 GB", int64(12 * gb / 10)},
		{"512 B", 512},
		{"100 KB", 100 * 1024},
		{"100 kB", 100 * 1024},
		{"2.5 TB", int64(25 * tb / 10)},
		{"", 0},
		{"badvalue MB", 0},
		{"100", 0},
		{"-1 MB", 0},
	}
	for _, tc := range tests {
		got := ParseHumanSize(tc.input)
		if got != tc.want {
			t.Errorf("ParseHumanSize(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

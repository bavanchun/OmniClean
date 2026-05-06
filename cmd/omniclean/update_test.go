package main

import "testing"

func TestIsUpdateAvailable(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
		wantErr bool
	}{
		{name: "older version", current: "0.7.0", latest: "0.8.0", want: true},
		{name: "same version", current: "0.8.0", latest: "0.8.0", want: false},
		{name: "current has v prefix", current: "v0.8.0", latest: "0.8.0", want: false},
		{name: "dev version can update", current: "dev", latest: "0.8.0", want: true},
		{name: "invalid latest version", current: "0.8.0", latest: "latest", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := isUpdateAvailable(tc.current, tc.latest)
			if (err != nil) != tc.wantErr {
				t.Fatalf("isUpdateAvailable() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("isUpdateAvailable() = %v, want %v", got, tc.want)
			}
		})
	}
}

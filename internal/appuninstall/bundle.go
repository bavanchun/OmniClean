//go:build darwin

package appuninstall

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type plistData struct {
	BundleID    string `json:"CFBundleIdentifier"`
	Version     string `json:"CFBundleShortVersionString"`
	DisplayName string `json:"CFBundleDisplayName"`
	BundleName  string `json:"CFBundleName"`
}

// ParseBundle reads Info.plist from appPath/Contents/Info.plist via plutil.
// On failure it returns a Bundle with Name derived from path and empty BundleID/Version.
func ParseBundle(ctx context.Context, appPath string) Bundle {
	name := strings.TrimSuffix(filepath.Base(appPath), ".app")
	b := Bundle{Path: appPath, Name: name}

	info, err := os.Stat(appPath)
	if err == nil {
		b.LastModTime = info.ModTime()
	}
	b.Size = dirSize(appPath)

	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	out, err := exec.CommandContext(ctx, "plutil", "-convert", "json", "-o", "-", plistPath).Output()
	if err != nil {
		return b
	}

	var data plistData
	if err := json.Unmarshal(out, &data); err != nil {
		return b
	}

	b.BundleID = data.BundleID
	b.Version = data.Version
	// prefer DisplayName, then BundleName, then filename
	if data.DisplayName != "" {
		b.Name = data.DisplayName
	} else if data.BundleName != "" {
		b.Name = data.BundleName
	}
	return b
}

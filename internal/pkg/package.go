// Package pkg defines the core data types used throughout OmniClean.
package pkg

import (
	"fmt"
	"strconv"
	"strings"
)

// ManagerType identifies which package manager owns a package.
type ManagerType string

const (
	ManagerAPT     ManagerType = "apt"
	ManagerBrew    ManagerType = "brew"
	ManagerSnap    ManagerType = "snap"
	ManagerFlatpak ManagerType = "flatpak"
	ManagerPip     ManagerType = "pip"
	ManagerNPM     ManagerType = "npm"
	ManagerCargo   ManagerType = "cargo"
	// Windows package managers
	ManagerWinget ManagerType = "winget"
	ManagerChoco  ManagerType = "choco"
	ManagerScoop  ManagerType = "scoop"
)

// Package represents a single installed package from any package manager.
// It implements the bubbles list.Item interface for direct use in the TUI list.
type Package struct {
	Name        string
	Version     string
	Manager     ManagerType
	Description string
	// Size in bytes, 0 if unknown
	Size int64
}

// FilterValue is used by bubbles/list for fuzzy filtering.
func (p Package) FilterValue() string {
	return p.Name
}

// Title returns the primary display string for the package in the list.
func (p Package) Title() string {
	return p.Name
}

// Desc returns the secondary display string for the package in the list.
func (p Package) Desc() string {
	if p.Description != "" {
		return fmt.Sprintf("[%s %s] %s", p.Manager, p.Version, p.Description)
	}
	return fmt.Sprintf("[%s] %s", p.Manager, p.Version)
}

// ParseHumanSize converts a human-readable size string (e.g. "56.6 MB", "1.2 GB")
// to bytes. Returns 0 if the string cannot be parsed.
func ParseHumanSize(s string) int64 {
	s = strings.TrimSpace(s)
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0
	}
	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || val < 0 {
		return 0
	}
	switch strings.ToUpper(parts[1]) {
	case "B":
		return int64(val)
	case "KB", "KIB":
		return int64(val * 1024)
	case "MB", "MIB":
		return int64(val * 1024 * 1024)
	case "GB", "GIB":
		return int64(val * 1024 * 1024 * 1024)
	case "TB", "TIB":
		return int64(val * 1024 * 1024 * 1024 * 1024)
	}
	return 0
}

// UninstallResult holds the outcome of a single package removal operation.
type UninstallResult struct {
	Package   Package
	Err       error
	DryRunCmd string   // set when dry-run mode; shows the command that would run
	Leftovers []string // leftover config/cache paths found after uninstall
}

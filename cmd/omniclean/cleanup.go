package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/bavanchun/OmniClean/internal/cleanup"
	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/pkg"
	cleanuptui "github.com/bavanchun/OmniClean/internal/tui/cleanup"
)

// cleanupRecord is the stable JSON shape for one removable candidate. It
// mirrors the analyze --json struct-tag convention (lowercase keys, omitempty
// for unknown values). InstalledAt is a pointer so a zero install time is
// omitted entirely rather than serialized as the Go zero date.
type cleanupRecord struct {
	Name        string     `json:"name"`
	Manager     string     `json:"manager"`
	Version     string     `json:"version"`
	Role        string     `json:"role"`
	InstalledAt *time.Time `json:"installedAt,omitempty"`
	Size        int64      `json:"size"`
}

// newCleanupCmd registers `omniclean cleanup`, surfacing the cleanup advisor for
// scripting. With --json (or a non-TTY stdout) it emits the removable-candidate
// list as JSON; otherwise it launches the interactive Cleanup Suggestions TUI.
func newCleanupCmd() *cobra.Command {
	var (
		jsonOut  bool
		managers []string
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Review and remove orphan/leaf packages",
		Long: `Aggregate removable cleanup candidates (orphaned dependencies and
top-level leaf packages) across all available managers. Pass --json or pipe the
output for a machine-readable list; otherwise an interactive TUI opens.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			detectors := detector.AvailableDetectors()
			if len(managers) > 0 {
				detectors = filterDetectors(detectors, managers)
			}
			if len(detectors) == 0 {
				fmt.Fprintln(os.Stderr, "No supported package managers found on this system.")
				return nil
			}

			if shouldEmitJSON(jsonOut) {
				candidates := cleanup.Aggregate(context.Background(), detectors)
				return writeCleanupJSON(os.Stdout, candidates)
			}

			app := cleanuptui.New(cleanuptui.Config{Detectors: detectors, DryRun: dryRun})
			return app.Run(context.Background())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON instead of opening the TUI")
	cmd.Flags().StringSliceVarP(&managers, "manager", "m", nil, "Filter to specific manager(s) (e.g. apt,brew)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Simulate removal without making changes")
	return cmd
}

// writeCleanupJSON serializes removable candidates to w with two-space indent,
// matching the analyze --json convention. Only RoleManual/RoleOrphan packages
// reach here (the aggregator filters), so role ∈ {manual, orphan}.
func writeCleanupJSON(w io.Writer, pkgs []pkg.Package) error {
	records := make([]cleanupRecord, 0, len(pkgs))
	for _, p := range pkgs {
		rec := cleanupRecord{
			Name:    p.Name,
			Manager: string(p.Manager),
			Version: p.Version,
			Role:    string(p.Role),
			Size:    p.Size,
		}
		if !p.InstalledAt.IsZero() {
			t := p.InstalledAt.UTC()
			rec.InstalledAt = &t
		}
		records = append(records, rec)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(records)
}

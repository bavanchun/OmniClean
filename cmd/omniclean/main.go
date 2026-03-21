// OmniClean — unified package manager uninstaller.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bavanchun/OmniClean/internal/detector"
	"github.com/bavanchun/OmniClean/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func main() {
	var (
		dryRun    bool
		managers  []string
		noConfirm bool
	)

	root := &cobra.Command{
		Use:     "omniclean",
		Short:   "Unified package manager uninstaller",
		Version: version,
		Long: `OmniClean aggregates packages from multiple package managers
into a single interactive TUI. Search, select, and cleanly uninstall.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			detectors := detector.AvailableDetectors()
			if len(managers) > 0 {
				detectors = filterDetectors(detectors, managers)
			}
			if len(detectors) == 0 {
				fmt.Fprintln(os.Stderr, "No supported package managers found on this system.")
				return nil
			}

			app := tui.New(tui.Config{
				Detectors: detectors,
				DryRun:    dryRun,
				NoConfirm: noConfirm,
			})
			return app.Run(context.Background())
		},
	}

	root.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Simulate uninstallation without making changes")
	root.Flags().StringSliceVarP(&managers, "manager", "m", nil, "Filter to specific manager(s) (e.g. apt,pip)")
	root.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompt before uninstalling")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func filterDetectors(all []detector.Detector, names []string) []detector.Detector {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	filtered := make([]detector.Detector, 0, len(all))
	for _, d := range all {
		if nameSet[d.Name()] {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

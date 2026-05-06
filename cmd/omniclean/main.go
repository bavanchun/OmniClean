// OmniClean — unified package manager uninstaller.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/bavanchun/OmniClean/internal/detector"
	logger "github.com/bavanchun/OmniClean/internal/logger"
	"github.com/bavanchun/OmniClean/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func main() {
	var (
		dryRun    bool
		managers  []string
		noConfirm bool
		verbose   bool
	)

	root := &cobra.Command{
		Use:     "omniclean",
		Short:   "Unified package manager uninstaller",
		Version: version,
		Long: `OmniClean aggregates packages from multiple package managers
into a single interactive TUI. Search, select, and cleanly uninstall.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Configure logging
			if verbose {
				logger.SetVerbose()
			}

			// Set up file logging for TUI mode (TUI owns stdout/stderr)
			cleanup := logger.SetupFileLogging()
			defer cleanup()

			logger.L.Info("starting omniclean",
				"version", version,
				"dry_run", dryRun,
				"verbose", verbose,
			)

			detectors := detector.AvailableDetectors()
			if len(managers) > 0 {
				detectors = filterDetectors(detectors, managers)
			}
			if len(detectors) == 0 {
				fmt.Fprintln(os.Stderr, "No supported package managers found on this system.")
				return nil
			}

			logger.L.Info("detected package managers",
				"count", len(detectors),
				"managers", detectorNames(detectors),
			)

			app := tui.New(tui.Config{
				Detectors: detectors,
				DryRun:    dryRun,
				NoConfirm: noConfirm,
				Verbose:   verbose,
			})
			return app.Run(context.Background())
		},
	}

	root.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Simulate uninstallation without making changes")
	root.Flags().StringSliceVarP(&managers, "manager", "m", nil, "Filter to specific manager(s) (e.g. apt,pip)")
	root.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompt before uninstalling")
	root.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug logging")

	root.AddCommand(newUpdateCmd())
	root.AddCommand(newPurgeCmd())
	root.AddCommand(newAnalyzeCmd())

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

func detectorNames(detectors []detector.Detector) []string {
	names := make([]string, len(detectors))
	for i, d := range detectors {
		names[i] = d.Name()
	}
	return names
}

// Ensure log package is used (it's imported for its side effects in the logger package).
var _ = log.DebugLevel

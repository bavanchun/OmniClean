// OmniClean — unified package manager uninstaller.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/bavanchun/OmniClean/internal/analyze"
	"github.com/bavanchun/OmniClean/internal/detector"
	logger "github.com/bavanchun/OmniClean/internal/logger"
	"github.com/bavanchun/OmniClean/internal/purge"
	"github.com/bavanchun/OmniClean/internal/tui"
	analyzetui "github.com/bavanchun/OmniClean/internal/tui/analyze"
	menutui "github.com/bavanchun/OmniClean/internal/tui/menu"
	purgetui "github.com/bavanchun/OmniClean/internal/tui/purge"
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
			if verbose {
				logger.SetVerbose()
			}
			cleanup := logger.SetupFileLogging()
			defer cleanup()

			sel, err := menutui.Run()
			if err != nil {
				return fmt.Errorf("menu: %w", err)
			}

			switch sel {
			case menutui.SelectUninstall:
				return runUninstall(context.Background(), dryRun, noConfirm, verbose, managers)
			case menutui.SelectAnalyze:
				return runAnalyzeDefault(context.Background())
			case menutui.SelectPurge:
				return runPurgeDefault(context.Background(), dryRun, noConfirm)
			default:
				return nil // Quit
			}
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

func runUninstall(ctx context.Context, dryRun, noConfirm, verbose bool, managers []string) error {
	detectors := detector.AvailableDetectors()
	if len(managers) > 0 {
		detectors = filterDetectors(detectors, managers)
	}
	if len(detectors) == 0 {
		fmt.Fprintln(os.Stderr, "No supported package managers found on this system.")
		return nil
	}
	logger.L.Info("detected package managers", "count", len(detectors), "managers", detectorNames(detectors))
	app := tui.New(tui.Config{
		Detectors: detectors,
		DryRun:    dryRun,
		NoConfirm: noConfirm,
		Verbose:   verbose,
	})
	return app.Run(ctx)
}

func runAnalyzeDefault(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	app := analyzetui.New(analyzetui.Config{
		Path:    abs,
		Options: analyze.Options{LargeFileTopN: 20, LargeFileMinBytes: 50 << 20},
	})
	return app.Run(ctx)
}

func runPurgeDefault(ctx context.Context, dryRun, noConfirm bool) error {
	roots, err := purge.LoadRoots(purge.DefaultPathsFile())
	if err != nil {
		return fmt.Errorf("load roots: %w", err)
	}
	app := purgetui.New(purgetui.Config{
		Roots:     roots,
		Options:   purge.Options{RecentDays: purge.DefaultRecentDays},
		DryRun:    dryRun,
		NoConfirm: noConfirm,
	})
	return app.Run(ctx)
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

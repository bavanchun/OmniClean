package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bavanchun/OmniClean/internal/analyze"
	analyzetui "github.com/bavanchun/OmniClean/internal/tui/analyze"
)

func newAnalyzeCmd() *cobra.Command {
	var (
		jsonOut   bool
		largeMin  string
		largeTopN int
	)

	cmd := &cobra.Command{
		Use:   "analyze [path]",
		Short: "Explore disk usage interactively",
		Long: `Analyze the given directory (default: $HOME) and present a TUI
disk explorer. Pass --json or pipe the output to switch to a
machine-readable JSON payload.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := defaultAnalyzePath()
			if len(args) == 1 {
				path = args[0]
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("resolve path: %w", err)
			}

			minBytes, err := parseSize(largeMin)
			if err != nil {
				return fmt.Errorf("--large-min: %w", err)
			}
			opts := analyze.Options{
				LargeFileTopN:     largeTopN,
				LargeFileMinBytes: minBytes,
			}

			if shouldEmitJSON(jsonOut) {
				res, err := analyze.NewWalker().Scan(context.Background(), abs, opts)
				if err != nil {
					return err
				}
				return analyze.WriteJSON(os.Stdout, res)
			}

			app := analyzetui.New(analyzetui.Config{Path: abs, Options: opts})
			return app.Run(context.Background())
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON instead of opening the TUI")
	cmd.Flags().StringVar(&largeMin, "large-min", "100MB", "size threshold for the large-files panel")
	cmd.Flags().IntVar(&largeTopN, "large-top", 20, "max number of large files to surface")

	return cmd
}

// shouldEmitJSON honors the --json flag and auto-detects when stdout is
// not a terminal (piped or redirected). The character-device test
// works across Linux, macOS, and Windows without extra dependencies.
func shouldEmitJSON(flag bool) bool {
	if flag {
		return true
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func defaultAnalyzePath() string {
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return h
	}
	return "."
}

// parseSize accepts forms like "100MB", "1GB", "500K" and returns the
// equivalent byte count. Empty string defaults to 100 MiB to match the
// scanner default. Errors stay terse so the cobra layer can show them.
func parseSize(s string) (int64, error) {
	if s == "" {
		return 100 << 20, nil
	}
	var n int64
	var unit string
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
			continue
		}
		unit += string(c)
	}
	switch unit {
	case "", "B":
		return n, nil
	case "K", "KB", "KiB":
		return n << 10, nil
	case "M", "MB", "MiB":
		return n << 20, nil
	case "G", "GB", "GiB":
		return n << 30, nil
	case "T", "TB", "TiB":
		return n << 40, nil
	}
	return 0, fmt.Errorf("unknown size unit %q", unit)
}

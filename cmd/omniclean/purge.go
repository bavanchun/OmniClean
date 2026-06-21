package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bavanchun/OmniClean/internal/purge"
	purgetui "github.com/bavanchun/OmniClean/internal/tui/purge"
)

func newPurgeCmd() *cobra.Command {
	var (
		dryRun     bool
		noConfirm  bool
		editPaths  bool
		stacks     []string
		recentDays int
	)

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Find and delete project artifact directories (node_modules, target, .venv, ...)",
		Long: `Scan configured project roots for ephemeral build artifacts and
let the user delete them after review. Honors ~/.config/omniclean/purge_paths
when present.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			roots, err := purge.LoadRoots(purge.DefaultPathsFile())
			if err != nil {
				return fmt.Errorf("load roots: %w", err)
			}

			opts := purge.Options{
				RecentDays:    recentDays,
				IncludeStacks: convertStacks(stacks),
			}

			if editPaths {
				return purgetui.EditPaths(purge.DefaultPathsFile(), roots)
			}

			app := purgetui.New(purgetui.Config{
				Roots:     roots,
				Options:   opts,
				DryRun:    dryRun,
				NoConfirm: noConfirm,
			})
			return app.Run(context.Background())
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview deletions without removing files")
	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "skip confirmation before deletion")
	cmd.Flags().BoolVar(&editPaths, "paths", false, "edit configured scan roots")
	cmd.Flags().StringSliceVarP(&stacks, "stack", "s", nil, "restrict to stacks (node,rust,python,go,java,dotnet,build)")
	cmd.Flags().IntVar(&recentDays, "recent-days", purge.DefaultRecentDays, "treat targets newer than N days as Recent (pre-unselected)")

	return cmd
}

func convertStacks(in []string) []purge.Stack {
	if len(in) == 0 {
		return nil
	}
	out := make([]purge.Stack, 0, len(in))
	for _, s := range in {
		out = append(out, purge.Stack(s))
	}
	return out
}

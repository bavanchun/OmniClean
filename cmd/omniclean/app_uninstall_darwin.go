//go:build darwin

package main

import (
	"context"

	appuninstalltui "github.com/bavanchun/OmniClean/internal/tui/appuninstall"
)

func runUninstallApps(ctx context.Context, dryRun bool) error {
	app := appuninstalltui.New(appuninstalltui.Config{DryRun: dryRun})
	return app.Run(ctx)
}

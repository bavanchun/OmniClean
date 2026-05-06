// Package purge wraps the bubbletea program that drives the
// `omniclean purge` command. The list/confirm/result views are filled
// in by later commits in Phase 2; this file declares the public seams
// so cmd/omniclean can import them today.
package purge

import (
	"context"
	"errors"

	"github.com/bavanchun/OmniClean/internal/purge"
)

// Config captures everything the cobra command needs to hand off to
// the TUI program.
type Config struct {
	Roots     []string
	Options   purge.Options
	DryRun    bool
	NoConfirm bool
}

// App is the placeholder Bubbletea program. Run currently returns a
// "not implemented" error so the binary builds end-to-end while the
// view files are added in subsequent commits.
type App struct{ cfg Config }

// New constructs the App from the supplied Config.
func New(cfg Config) *App { return &App{cfg: cfg} }

// Run starts the TUI. The list/confirm/result wiring lands in Phase 2.7
// and 2.8.
func (a *App) Run(ctx context.Context) error {
	return errors.New("purge TUI is under construction; see roadmap Phase 2")
}

// EditPaths opens the interactive paths editor. Implementation lands in
// Phase 2.9.
func EditPaths(_ string, _ []string) error {
	return errors.New("purge --paths editor is under construction; see roadmap Phase 2.9")
}

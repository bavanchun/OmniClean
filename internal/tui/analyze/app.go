// Package analyze drives the interactive disk explorer used by
// `omniclean analyze`. The skeleton here makes the binary build today;
// list/navigation/large-files/trash views land in Phase 3.7-3.10.
package analyze

import (
	"context"
	"errors"

	"github.com/bavanchun/OmniClean/internal/analyze"
)

// Config wires the cobra command to the TUI program.
type Config struct {
	Path    string
	Options analyze.Options
}

// App is the placeholder Bubbletea program.
type App struct{ cfg Config }

// New constructs the App from the supplied Config.
func New(cfg Config) *App { return &App{cfg: cfg} }

// Run starts the TUI. The interactive views land in subsequent commits.
func (a *App) Run(ctx context.Context) error {
	return errors.New("analyze TUI is under construction; see roadmap Phase 3.7-3.10")
}

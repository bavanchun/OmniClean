package purge

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"

	corepurge "github.com/bavanchun/OmniClean/internal/purge"
)

// EditPaths opens an interactive editor that lets the user revise the
// configured purge scan roots. On submit the new list is written back
// to configPath via corepurge.SaveRoots. configPath being empty means
// the OS could not provide a config dir; in that case we tell the user
// instead of silently dropping their edits.
func EditPaths(configPath string, current []string) error {
	if configPath == "" {
		return fmt.Errorf("no user config directory available; cannot persist purge paths")
	}

	value := strings.Join(current, "\n")
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Purge scan roots").
				Description("One path per line. Use ~ for $HOME. Lines starting with # are ignored.").
				Lines(12).
				CharLimit(8192).
				Value(&value),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	roots := splitNonEmpty(value)
	if err := corepurge.SaveRoots(configPath, roots); err != nil {
		return fmt.Errorf("save roots: %w", err)
	}
	fmt.Printf("Saved %d path(s) to %s\n", len(roots), configPath)
	return nil
}

// splitNonEmpty trims and drops blank/comment lines so the persisted
// file matches what LoadRoots will accept.
func splitNonEmpty(blob string) []string {
	var out []string
	for _, raw := range strings.Split(blob, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

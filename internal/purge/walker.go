package purge

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewWalker returns a default Scanner backed by filepath.WalkDir.
// The fd-backed acceleration path is intentionally deferred to a later
// commit; this walker is portable and does not require any external
// binary, so it works on every supported OS out of the box.
func NewWalker() Scanner { return &walker{now: time.Now} }

type walker struct {
	now func() time.Time
}

func (w *walker) Scan(ctx context.Context, roots []string, opts Options) ([]Target, error) {
	if opts.RecentDays == 0 {
		opts.RecentDays = 7
	}
	stackFilter := stackSet(opts.IncludeStacks)

	var targets []Target
	for _, root := range roots {
		if root == "" {
			continue
		}
		base := os.ExpandEnv(expandHome(root))
		if _, err := os.Stat(base); err != nil {
			continue
		}
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable subtree
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !d.IsDir() {
				return nil
			}
			// Bound depth by counting separators relative to base.
			if opts.MaxDepth > 0 && depth(base, path) > opts.MaxDepth {
				return fs.SkipDir
			}
			pat := matchPattern(d.Name())
			if pat == nil {
				return nil
			}
			if stackFilter != nil && !stackFilter[pat.Stack] {
				return fs.SkipDir
			}
			t := newTarget(path, *pat, w.now(), opts.RecentDays)
			targets = append(targets, t)
			// Do not descend into matched artifact dirs.
			return fs.SkipDir
		})
		if err != nil && ctx.Err() == nil {
			return targets, err
		}
	}
	return targets, nil
}

func newTarget(path string, pat Pattern, now time.Time, recentDays int) Target {
	size := dirSize(path)
	mod := time.Time{}
	if info, err := os.Stat(path); err == nil {
		mod = info.ModTime()
	}
	recent := false
	if !mod.IsZero() {
		recent = now.Sub(mod) < time.Duration(recentDays)*24*time.Hour
	}
	return Target{
		Path:     path,
		Project:  filepath.Base(filepath.Dir(path)),
		Stack:    pat.Stack,
		Pattern:  pat.Name,
		Size:     size,
		Modified: mod,
		Recent:   recent,
	}
}

func depth(base, path string) int {
	rel, err := filepath.Rel(base, path)
	if err != nil || rel == "." {
		return 0
	}
	return 1 + strings.Count(rel, string(filepath.Separator))
}

// dirSize sums the file sizes under path with a generous cap to avoid
// stalling on huge artifact dirs. Errors mid-walk are swallowed so we
// always return a best-effort total.
func dirSize(path string) int64 {
	const maxEntries = 200_000
	var total int64
	var entries int
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		entries++
		if entries > maxEntries {
			return fs.SkipAll
		}
		if d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

func stackSet(stacks []Stack) map[Stack]bool {
	if len(stacks) == 0 {
		return nil
	}
	m := make(map[Stack]bool, len(stacks))
	for _, s := range stacks {
		m[s] = true
	}
	return m
}

func expandHome(path string) string {
	if len(path) >= 2 && path[0] == '~' && (path[1] == '/' || path[1] == filepath.Separator) {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

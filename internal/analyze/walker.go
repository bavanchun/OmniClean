package analyze

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// NewWalker returns a Scanner that lists immediate children of the
// target directory and recursively totals each child's size in
// parallel. Large files surfaced via Result.LargeFiles come from a
// separate streaming pass capped at Options.LargeFileTopN.
func NewWalker() Scanner { return &walker{} }

type walker struct{}

func (w *walker) Scan(ctx context.Context, path string, opts Options) (Result, error) {
	if opts.MaxWorkers <= 0 {
		opts.MaxWorkers = runtime.NumCPU()
	}
	if opts.LargeFileTopN <= 0 {
		opts.LargeFileTopN = 20
	}
	if opts.LargeFileMinBytes <= 0 {
		opts.LargeFileMinBytes = 100 << 20 // 100 MiB
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return Result{Path: path}, err
	}

	res := Result{Path: path}
	var totalFiles atomic.Int64
	var totalSize atomic.Int64

	// Worker pool to size each immediate child concurrently.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(opts.MaxWorkers)
	dirs := make([]DirEntry, len(entries))
	for i, ent := range entries {
		i, ent := i, ent
		full := filepath.Join(path, ent.Name())
		g.Go(func() error {
			info, err := os.Lstat(full)
			if err != nil {
				return nil
			}
			d := DirEntry{
				Name:       ent.Name(),
				Path:       full,
				IsDir:      ent.IsDir(),
				LastAccess: info.ModTime(),
			}
			if ent.IsDir() {
				size, files := dirTotals(gctx, full)
				d.Size = size
				totalFiles.Add(files)
			} else {
				d.Size = info.Size()
				totalFiles.Add(1)
			}
			totalSize.Add(d.Size)
			dirs[i] = d
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return res, err
	}

	// Drop empty placeholders from skipped lstat errors.
	res.Entries = make([]DirEntry, 0, len(dirs))
	for _, d := range dirs {
		if d.Path != "" {
			res.Entries = append(res.Entries, d)
		}
	}
	sort.Slice(res.Entries, func(i, j int) bool {
		return res.Entries[i].Size > res.Entries[j].Size
	})

	res.TotalSize = totalSize.Load()
	res.TotalFiles = totalFiles.Load()
	res.LargeFiles = collectLargeFiles(ctx, path, opts)
	return res, nil
}

// dirTotals walks dir under ctx and sums sizes plus file counts.
// Errors mid-walk are swallowed so a single denied subdir does not
// abort the whole scan.
func dirTotals(ctx context.Context, dir string) (int64, int64) {
	var size, files int64
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return nil
		}
		size += info.Size()
		files++
		return nil
	})
	return size, files
}

// collectLargeFiles streams the tree once collecting the top-N files
// by size whose size meets opts.LargeFileMinBytes. We use a small heap-
// less approach: keep a sorted slice trimmed to TopN.
func collectLargeFiles(ctx context.Context, root string, opts Options) []FileEntry {
	var (
		mu  sync.Mutex
		top []FileEntry
	)
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || ctx.Err() != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil || info.Size() < opts.LargeFileMinBytes {
			return nil
		}
		entry := FileEntry{Name: d.Name(), Path: p, Size: info.Size()}
		mu.Lock()
		top = append(top, entry)
		sort.Slice(top, func(i, j int) bool { return top[i].Size > top[j].Size })
		if len(top) > opts.LargeFileTopN {
			top = top[:opts.LargeFileTopN]
		}
		mu.Unlock()
		return nil
	})
	return top
}

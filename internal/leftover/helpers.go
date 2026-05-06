package leftover

import "github.com/bavanchun/OmniClean/internal/pkg"

// scanCandidates is the boilerplate every per-manager scanner runs after
// it builds a list of candidate paths: filter to existing entries, size
// each, and check the whitelist. Centralising it here keeps the
// per-manager files focused on path math.
func scanCandidates(mgr pkg.ManagerType, name string, candidates []string, w *Whitelist) Result {
	res := Result{Manager: mgr, Package: name}
	for _, p := range existing(candidates) {
		size, _ := pathSize(p, SizeLimits{})
		res.Entries = append(res.Entries, Entry{
			Path:        p,
			Size:        size,
			Whitelisted: w.Match(p),
		})
		res.Total += size
	}
	return res
}

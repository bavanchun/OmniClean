package leftover

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Whitelist holds glob patterns that mark leftover paths as protected.
// Patterns are matched with filepath.Match against the absolute path,
// with $HOME expansion done at load time.
type Whitelist struct {
	patterns []string
}

// LoadWhitelist reads patterns from path. Missing files yield an empty
// whitelist, not an error, so first-run users get sensible defaults
// without touching the filesystem.
func LoadWhitelist(path string) (*Whitelist, error) {
	w := &Whitelist{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return w, nil
		}
		return nil, err
	}
	defer f.Close()

	home, _ := os.UserHomeDir()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "~/") && home != "" {
			line = filepath.Join(home, strings.TrimPrefix(line, "~/"))
		}
		w.patterns = append(w.patterns, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return w, nil
}

// Match returns true when any whitelist pattern matches the given path.
// An empty whitelist always returns false. Match errors from
// filepath.Match (e.g. malformed patterns) are treated as non-matches so
// a single bad entry never widens the protected set.
func (w *Whitelist) Match(path string) bool {
	if w == nil {
		return false
	}
	for _, p := range w.patterns {
		ok, err := filepath.Match(p, path)
		if err == nil && ok {
			return true
		}
		// Allow prefix-style entries: "/foo/bar" matches "/foo/bar/...".
		if strings.HasPrefix(path, strings.TrimSuffix(p, "/*")+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// DefaultWhitelistPath returns the conventional whitelist file location
// inside the user's config directory. Empty when no config dir is
// available (rare; usually only on misconfigured systems).
func DefaultWhitelistPath() string {
	cfg, err := os.UserConfigDir()
	if err != nil || cfg == "" {
		return ""
	}
	return filepath.Join(cfg, "omniclean", "whitelist")
}

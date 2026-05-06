package purge

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultRoots are the directories scanned when the user has not
// configured custom paths. Tilde expansion happens inside walker.
var DefaultRoots = []string{
	"~/Projects",
	"~/Code",
	"~/dev",
	"~/Developer",
	"~/workspace",
	"~/source/repos", // Visual Studio default on Windows
	"~/GitHub",
}

// DefaultPathsFile returns the conventional config path for the
// user-configured scan roots. Empty when no config dir is available.
func DefaultPathsFile() string {
	cfg, err := os.UserConfigDir()
	if err != nil || cfg == "" {
		return ""
	}
	return filepath.Join(cfg, "omniclean", "purge_paths")
}

// LoadRoots reads scan roots from path. Missing files yield DefaultRoots
// so first-run users still get sensible coverage. Lines starting with
// '#' and blank lines are ignored.
func LoadRoots(path string) ([]string, error) {
	if path == "" {
		return append([]string(nil), DefaultRoots...), nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return append([]string(nil), DefaultRoots...), nil
		}
		return nil, err
	}
	defer f.Close()

	var roots []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		roots = append(roots, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return append([]string(nil), DefaultRoots...), nil
	}
	return roots, nil
}

// SaveRoots writes roots to path, creating parent dirs as needed.
// One path per line; the file is rewritten atomically through a temp
// file so a crash cannot leave a half-written config.
func SaveRoots(path string, roots []string) error {
	if path == "" {
		return fmt.Errorf("empty config path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "purge_paths.*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	w := bufio.NewWriter(tmp)
	for _, r := range roots {
		fmt.Fprintln(w, r)
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

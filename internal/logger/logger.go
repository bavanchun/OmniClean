// Package logger provides a centralized logging solution using charmbracelet/log.
// When the TUI is active, logs are written to a file; otherwise to stderr.
package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

// Global logger instance.
var L *log.Logger

func init() {
	L = log.NewWithOptions(os.Stderr, log.Options{
		Prefix:          "omniclean",
		ReportTimestamp: true,
		Level:           log.InfoLevel,
	})
}

// SetLevel sets the global log level.
func SetLevel(level log.Level) {
	L.SetLevel(level)
}

// SetVerbose enables debug-level logging.
func SetVerbose() {
	L.SetLevel(log.DebugLevel)
}

// SetupFileLogging redirects log output to a file at
// ~/.config/omniclean/omniclean.log while keeping stderr for fatal errors.
// Returns a cleanup function to close the file.
func SetupFileLogging() func() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		L.Warn("cannot determine config dir, logging to stderr", "err", err)
		return func() {}
	}

	logDir := filepath.Join(configDir, "omniclean")
	if mkErr := os.MkdirAll(logDir, 0o755); mkErr != nil {
		L.Warn("cannot create log directory", "err", mkErr)
		return func() {}
	}

	logPath := filepath.Join(logDir, "omniclean.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		L.Warn("cannot open log file", "path", logPath, "err", err)
		return func() {}
	}

	// Write logs to both file and stderr for fatal errors
	L.SetOutput(io.MultiWriter(f))
	L.Debug("log file opened", "path", logPath)

	return func() {
		_ = f.Close()
	}
}

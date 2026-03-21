# Contributing to OmniClean

Thanks for your interest in contributing! Here are the guidelines.

## Getting Started

1. Fork the repository and create a branch from `main`
2. Make your changes
3. Run `make test && make lint` — both must pass
4. Submit a pull request

## Adding a New Package Manager

OmniClean uses a [Detector pattern](internal/detector/detector.go) — each package manager is one file that implements the `Detector` interface.

### Steps

1. **Create** `internal/detector/<name>.go` implementing all interface methods:
   ```go
   type Detector interface {
       Name() string
       Available() bool         // checks if the binary exists via LookPath
       NeedsSudo() bool         // true if uninstall requires elevated privileges
       DryRunCommand(p pkg.Package) string  // command string for --dry-run display
       UninstallExecCmd(p pkg.Package) *exec.Cmd  // non-nil only if NeedsSudo()=true
       ListPackages(ctx context.Context) ([]pkg.Package, error)
       Uninstall(ctx context.Context, p pkg.Package) error
   }
   ```
2. **Write tests** in `internal/detector/<name>_test.go` — use a mock `CommandRunner` to avoid requiring the binary in CI
3. **Register** the detector in `internal/detector/registry.go`'s `AllDetectors()` function
4. **Add a `ManagerType`** constant in `internal/pkg/package.go`
5. **Add a badge color** in `internal/tui/styles.go`'s `ManagerBadge` map

### Notes

- Use `DefaultRunner` for production; inject a mock runner in tests
- If `NeedsSudo()` is `true`, implement `UninstallExecCmd` to return an `*exec.Cmd` with `Stdin/Stdout/Stderr` connected — the TUI will use `tea.ExecProcess` to give it terminal access
- If `NeedsSudo()` is `false`, `UninstallExecCmd` should return `nil`
- Never call `fmt.Printf` or write to stdout/stderr in a detector — it corrupts the TUI's AltScreen

## Commit Style

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add winget detector for Windows
fix: guard snap list against empty output
test: add table-driven tests for apt parser
docs: update README with go install instructions
```

## Code Style

- Wrap errors with context: `fmt.Errorf("apt list: %w", err)`
- Table-driven tests with `t.Run`
- No exported types from `internal/` packages

## Running Tests

```bash
make test      # runs go test -race ./... with coverage
make lint      # runs golangci-lint
```

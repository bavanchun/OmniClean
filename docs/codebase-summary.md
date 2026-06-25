# OmniClean Codebase Summary

## Project Overview

OmniClean is a cross-platform Go TUI application providing unified cleanup for package managers, project artifacts, and disk space management.

**Key Characteristics:**
- **Language:** Go 1.21+
- **UI Framework:** Bubbletea (TUI), Lipgloss (styling)
- **CLI Framework:** Cobra
- **Platforms:** macOS, Linux, Windows (with platform-specific features)

## Core Architecture

### Package Structure

#### `cmd/omniclean/`
Main entry point. Cobra CLI wires the main menu to four features:
- `SelectUninstall`: Package manager uninstall
- `SelectAnalyze`: Disk explorer
- `SelectPurge`: Project artifact purge
- `SelectUninstallApps`: macOS .app bundle uninstall (Darwin only)

#### `internal/tui/`
Bubbletea TUI implementations for all features:
- **menu/** — Feature selection screen with animated effects (`--fancy` flag)
- **appuninstall/** — macOS app bundle deletion flow (10-state FSM)
- **analyze/** — Disk explorer with viewport scrolling
- **purge/** — Project artifact cleanup with delete state
- **components/** — Reusable UI widgets (badges, panels, keyhints)
- **theme/** — Color tokens and style definitions

#### `internal/appuninstall/` (Darwin only)
Core logic for macOS app uninstall:
- **scanner.go** — Scans /Applications and ~/Applications, walks .app bundles
- **bundle.go** — Plist parsing for Bundle metadata (Name, BundleID, Version)
- **leftover.go** — Finds orphan files in Library/Preferences, Library/Caches, etc.
- **uninstaller.go** — Executes rm -rf on bundles and leftovers
- **types.go** — Bundle, LeftoverEntry, DeleteResult definitions

#### `internal/analyze/`
Disk space analysis:
- **scanner.go** — Recursive directory traversal with size calculation
- **walker.go** — Entry history and navigation state
- **large_files.go** — Top-N large files detection
- **trash.go** — OS-native trash integration (darwin/linux/windows)

#### `internal/purge/`
Project artifact management:
- **scanner.go** — Scans roots for artifact patterns (node_modules, target, .venv, etc.)
- **patterns.go** — Language/framework-specific artifact definitions
- **config.go** — User-editable scan root configuration
- **age.go** — Recent-build protection (mtime-based filtering)

#### `internal/detector/`
Package manager detection and listing:
- Per-manager files: apt.go, brew.go, cargo.go, flatpak.go, npm.go, pip.go, snap.go, winget.go, choco.go, scoop.go
- **detector.go** — Registry and cross-platform orchestration
- Platform-specific registry scanners (Windows)

#### `internal/leftover/`
Orphan file detection across package managers:
- **scanner.go** — Orchestrates per-manager whitelist + size calculation
- Manager-specific cleaners: apt.go, brew.go, cargo.go, flatpak.go, npm.go, pip.go, snap.go, scoop.go, choco.go, winget.go
- **whitelist.go** — User-editable ~/.config/omniclean/whitelist

#### `internal/cleaner/`
Package uninstall execution:
- Subprocess orchestration with dry-run support

#### `internal/logger/`
Structured logging with file rotation:
- Verbose mode support

## Feature Matrix

| Feature | Status | Platforms | Entry Point |
|---------|--------|-----------|-------------|
| Package uninstall | Complete | Linux, macOS, Windows | `internal/tui/`, menu selection 1 |
| Disk analyze | Complete | Linux, macOS, Windows | `internal/tui/analyze/`, menu selection 2 |
| Project purge | Complete | Linux, macOS, Windows | `internal/tui/purge/`, menu selection 3 |
| macOS app uninstall | Complete (Darwin) | macOS only | `internal/appuninstall/`, menu selection 4 |

## Recent Implementation (May 2026)

### 1. Analyze Disk TUI Improvements (`internal/tui/analyze/app.go`)
- **Viewport scrolling** with offset clamping for large entry lists
- **safeWidth() guard** for width calculations before first WindowSizeMsg
- **Large files overlay** with independent scroll state

### 2. Purge TUI Enhancements (`internal/tui/purge/app.go`)
- **Error state** handling for scan failures
- **Spinner during delete** for visual feedback
- **Footer text fix** for clarity
- **safeWidth() guard** for safe layout before size message

### 3. macOS App Uninstall Core (`internal/appuninstall/`)
**New packages:** Core logic for detecting and removing .app bundles
- **scanner.go** — Walks /Applications and ~/Applications at depth=1, parses .app bundle metadata
- **bundle.go** — Info.plist parser extracting Name, BundleID, Version, Size
- **leftover.go** — Orphan file finder in ~/Library/Preferences, ~/Library/Caches, ~/Library/Application Support
- **uninstaller.go** — rm -rf executor with DeleteResult tracking
- **types.go** — Bundle, LeftoverEntry, DeleteResult type definitions

### 4. macOS App Uninstall TUI (`internal/tui/appuninstall/`)
**10-state Bubbletea FSM:**
1. **stateLoading** — Scanning .app bundles
2. **stateList** — List view with selection checkboxes
3. **stateDetail** — Single app detail (populates leftover list on first access)
4. **stateConfirmBundle** — Confirm bundle deletion
5. **stateDeletingBundle** — Bundle removal in progress (spinner)
6. **stateLeftoverScan** — Auto-scan for orphan files post-bundle-delete
7. **stateConfirmLeftovers** — Confirm orphan cleanup
8. **stateDeletingLeftovers** — Leftover removal in progress
9. **stateResult** — Summary screen (deleted bundles + leftovers)
10. **stateError** — Unrecoverable error display

### 5. Menu Wiring (`internal/tui/menu/menu.go`, `cmd/omniclean/main.go`)
- **[4] Uninstall Apps** entry added to menu on Darwin only
- **runtime.GOOS** guard ensures macOS-only build stubs on other platforms
- Main command routes `SelectUninstallApps` → `runUninstallApps(ctx, dryRun)`

## Supported Package Managers

| Manager | Platforms | Detector | Leftover Scanner | Uninstall |
|---------|-----------|----------|-----------------|-----------|
| apt/dpkg | Linux (Debian/Ubuntu) | Yes | Yes | Yes |
| brew | macOS, Linux | Yes | Yes | Yes |
| snap | Linux | Yes | Yes | Yes |
| flatpak | Linux | Yes | Yes | Yes |
| pip | Cross-platform | Yes | Yes | Yes |
| npm | Cross-platform | Yes | Yes | Yes |
| cargo | Cross-platform | Yes | Yes | Yes |
| winget | Windows | Yes | Yes | Yes |
| choco | Windows | Yes | Yes | Yes |
| scoop | Windows | Yes | Yes | Yes |

## Key Conventions

### Error Handling
- Functions return `(result, error)` following Go idioms
- Wrapped with context: `fmt.Errorf("operation: %w", err)`
- TUI maps errors to `stateError` for user feedback

### State Machine Pattern
All TUI apps (analyze, purge, appuninstall) use `viewState` const + switch in `Update()`:
```go
type viewState int
const (
    stateLoading viewState = iota
    // ... other states
)

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    // ... handle messages
    }
    switch a.state {
    case stateLoading:
        // render loading view
    case stateList:
        // render list view
    // ...
    }
}
```

### Async Operations
Bubbletea Cmd/Msg pattern for async work:
- Long-running operations (scan, delete) return a `Cmd`
- Completion delivered via custom message type (e.g., `scanDoneMsg`)
- Spinner shown during operation with `spinner.TickMsg` polling

### Viewport Scrolling
- **scrollOffset** tracks render position in list
- **safeWidth()** guard for window size before first message arrives
- Offset clamping prevents overshooting list bounds

### Platform Guards
- **//go:build darwin** in appuninstall package files
- **runtime.GOOS == "darwin"** check in menu for conditional menu items
- Build stubs for non-Darwin platforms prevent import errors

## Configuration & Persistence

- **~/.config/omniclean/whitelist** — User-editable leftover exclusion list (JSON)
- **~/.config/omniclean/purge-roots.json** — User-configured scan directories for project artifacts

## Testing

- Unit tests in `*_test.go` files throughout internal packages
- Focus areas: scanner logic, pattern matching, whitelist parsing
- Coverage: ~70% (see `coverage.out`)

## Build & Deployment

- **Makefile** — `make build`, `make test`, `make lint`, `make fmt`, `make run`
- **GoReleaser** — Cross-platform binary builds + archives
- **GitHub Actions** — CI pipeline (golangci-lint, tests, coverage)
- **Latest Release** — Available at https://github.com/bavanchun/OmniClean/releases

## Known Limitations & TODOs

- App uninstall feature Darwin-only (macOS .app bundles)
- Large-files overlay in analyze TUI does not yet support deletion (planned)
- Whitelist editor is static JSON; UI editor planned for future release

# TUI Fixes + macOS App Uninstall Feature Complete

**Date**: 2026-05-07 00:54
**Severity**: Medium
**Component**: TUI (analyze, purge, appuninstall) + Menu
**Status**: Resolved

## What Happened

Completed 5-phase implementation: fixed Analyze & Purge TUI viewport crashes, built macOS app uninstaller (core + TUI), wired menu. 11 packages pass, 6 commits shipped.

## The Brutal Truth

This was a solid, methodical execution. Three annoyances: (1) safeWidth() infinite recursion caught mid-test — typo on recursive call, (2) spinner state machine was brittle with bad ordering, (3) platform-specific code forced runtime checks instead of build tags (menu file anatomy constraint).

No major delays. Clean delivery.

## Technical Details

**Analyze Disk TUI**: Added `scrollOffset` + `largeScrollOffset` with clamping. Guard `safeWidth()` against negative panel widths before first `WindowSizeMsg`. Reset scroll on directory nav.

**Purge TUI**: Added `stateError` field + `viewError()` to show error screen on scan failure. Batched `spinner.Tick()` with `startDelete()`. Fixed infinite recursion in `safeWidth()` (was calling itself). Stripped dev note from footer.

**App Uninstall Core** (Darwin-only):
- `scanner.go`: Scans /Applications + ~/Applications for .app bundles
- `bundle.go`: Parses Info.plist via `plutil -convert json`
- `leftover.go`: Probes 8 ~/Library locations for orphan files
- `uninstaller.go`: Deletes via osascript Finder trash, fallback to `os.RemoveAll`
- 7 unit tests, 60.4% coverage

**App Uninstall TUI**: 10-state Bubbletea flow (Loading → List → Detail → ConfirmBundle → DeletingBundle → LeftoverScan → ConfirmLeftovers → DeletingLeftovers → Result/Error). Split views.go from app.go.

**Menu**: Added `SelectUninstallApps` const, runtime guard `runtime.GOOS == "darwin"`, split into `app_uninstall_darwin.go` + `app_uninstall_other.go`. [4] appears macOS only.

## What We Tried

Safewidth() recursive bug caught during unit test phase — took 10 mins to spot typo.

Spinner animation only worked after batching with delete start (initial approach: separate goroutine, too much state collision).

## Root Cause Analysis

TUI crashes stemmed from assuming positive panel widths before layout settled. Lesson: guard all width/height math against uninitialized state.

Safewidth() typo was copy-paste error — test coverage saved us.

Spinner state needed synchronous ordering: tick must pair with action start, not race with goroutines.

## Lessons Learned

1. **Guard early, guard often**: Negative dimensions in TUI layout kill rendering. Add bounds checks before any math.
2. **Test catches silly mistakes**: Infinite recursion would ship without unit tests.
3. **Platform specificity via files, not tags**: Menu file stays generic; platform logic lives in separate files.
4. **State machines need synchronous events**: Spinner + delete state must sync, not async race.

## Next Steps

None blocking. App uninstall ready for feature testing. Monitor scanner edge cases (symlinks, permissions) in real usage.

**Files**: /internal/tui/analyze/app.go | /internal/tui/purge/app.go | /internal/appuninstall/* | /internal/tui/appuninstall/* | /internal/menu/menu.go | app_uninstall_darwin.go | app_uninstall_other.go

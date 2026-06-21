---
phase: 3
title: "Cleanup aggregation and TUI mode"
status: completed
priority: P1
effort: "L"
dependencies: [1, 2]
---

# Phase 3: Cleanup aggregation and TUI mode

## Overview

Aggregate classified packages across available detectors and surface them in a new "Cleanup Suggestions" TUI mode launched from the main menu. The mode shows only `Orphan` (first) and `Manual` packages, sorted by age (oldest `InstalledAt` first, unknown last), with role badges, and reuses the existing confirm + leftover-scan + delete flow.

## Requirements

- Functional: new menu entry → cleanup view listing removable candidates with role badge; selecting + confirming deletes via a self-contained confirm/delete flow.
- Non-functional: `Dependency` packages never appear as removable; `Unknown` packages from no-signal managers are excluded from suggestions (default exclude); no regression to existing modes; aggregation must not block the UI thread.

## Architecture

- **Aggregation** (NEW package `internal/cleanup` — NOT `internal/cleaner`, which is the uninstall executor with a different single responsibility): for each `AvailableDetectors()`, `ListPackages` then `ClassifyIfSupported` (Phase 1 helper); collect, filter to `RoleManual`/`RoleOrphan`, sort (Orphan before Manual, then oldest `InstalledAt` first, zero/unknown last).
  - Wrap each detector's `ListPackages`+`Classify` in a **bounded `context.WithTimeout`** so a slow/hung manager query degrades to whatever it returned (or skip) instead of freezing aggregation. (Red-team: `brew autoremove -n` is not guaranteed instant.)

- **Menu wiring** (`internal/tui/menu/menu.go`) — CRITICAL ordinal correctness. Selection is computed **positionally**: `menu.go:208` `Selection(a.cursor + 1)` and `:213` `Selection(idx + 1)`. The enum value equals the runtime item index + 1. `SelectQuit` is NOT produced positionally (set only by the quit key at `:220`), so it is a meaningless ordering anchor.
  - **Enum order must mirror `buildMenuItems` order for BOTH platform branches.** The darwin-only "Uninstall Apps" item is *appended last*, so `SelectUninstallApps` must remain the highest non-quit value.
  - **Exact change:** insert `SelectCleanup` immediately AFTER `SelectPurge` and BEFORE `SelectUninstallApps` in the enum; insert "Cleanup Suggestions" into the BASE (cross-platform) slice in `buildMenuItems` AFTER "Purge Project Artifacts" and BEFORE the `runtime.GOOS == "darwin"` append. Resulting index map:
    - non-darwin items `[Uninstall, Analyze, Purge, Cleanup]` → Selection 1..4 (`SelectUninstall`..`SelectCleanup`).
    - darwin items `[Uninstall, Analyze, Purge, Cleanup, UninstallApps]` → Selection 1..5; `SelectUninstallApps` stays the last/highest. ✓
  - A unit test MUST assert the cursor-index→Selection mapping for both `runtime.GOOS` branches (table test) to lock this invariant.

- **Command wiring** (`cmd/omniclean/main.go`): add `case menutui.SelectCleanup: return runCleanup(...)`; implement `runCleanup` mirroring `runUninstall` INCLUDING the `--manager` filter (`filterDetectors`, main.go:85) and dry-run.

- **TUI** (`internal/tui/cleanup/` new package): **self-contained**, following the shipped `internal/tui/appuninstall/` precedent — which does NOT import `internal/tui` (those `listModel`/`confirmModel`/delete message types are unexported inside the monolithic `App`). Reimplement a thin list + confirm + delete + async-loading state machine here, mirroring `appuninstall/{app,views,messages}.go`. Includes:
  - async loading via goroutine+channel (`appuninstall` `startLoading` precedent, app.go:156) — never call `cleanup.Aggregate` synchronously in `Init`/`Update`.
  - sudo sequencing for apt: reuse `detector.NewSudoExecCmd` + `tea.Exec` pattern (same mechanism the uninstall `App` uses); apt removal needs sudo (`apt.go:24`).
  - role badge column ("orphan"/"leaf").

## Related Code Files

- Create: `internal/cleanup/aggregate.go` (+ `aggregate_test.go`), `internal/tui/cleanup/{app,views,messages}.go` (+ test)
- Modify: `internal/tui/menu/menu.go` (enum ordinal + base-list item), `internal/tui/menu/menu_test.go` (new item snapshot + ordinal-mapping table test), `cmd/omniclean/main.go` (case + `runCleanup` with `--manager`/dry-run)
- Reference precedent (read, do NOT import): `internal/tui/appuninstall/*.go` (self-contained TUI), `internal/tui/list.go`/`confirm.go`/`app.go` (unexported — study, reimplement)

## Implementation Steps (TDD)

1. **Aggregate test first** — fake detectors (one classifier returning Manual/Orphan/Dependency, one non-classifier): assert output keeps only Manual/Orphan, Orphan first, then `InstalledAt` asc, Unknown excluded. Add a slow-fake case asserting timeout degrades gracefully.
2. Implement `cleanup.Aggregate(ctx, detectors)` with per-detector bounded timeout; green.
3. **Menu ordinal test first** — table test over both `runtime.GOOS` values asserting item-index→Selection mapping and that "Cleanup Suggestions" present, `SelectUninstallApps` stays last on darwin. Implement enum + base-list insertion; green.
4. Wire `runCleanup` (with `--manager` + dry-run) in `main.go`; build.
5. Implement self-contained cleanup TUI (async load + confirm + delete + apt sudo), mirroring `appuninstall`; manual smoke via `go run`.
6. `go test ./...` → green; `go vet ./...` clean.

## Success Criteria

- [x] `cleanup.Aggregate` filters to Manual/Orphan only, Orphan-first, age-sorted; slow detector degrades via timeout — all covered by tests.
- [x] Menu ordinal-mapping table test passes for darwin AND non-darwin; "Cleanup Suggestions" shown on all platforms; `SelectUninstallApps` remains last on darwin.
- [x] Cleanup TUI is self-contained (no import of unexported `internal/tui` symbols), lists candidates with role badge, async loading, dry-run honored, apt sudo sequenced.
- [x] No `Dependency` ever shown as removable.
- [x] Existing menu/uninstall/analyze/purge behavior unchanged; `go test ./...` green.

## Risk Assessment

- Risk: positional enum/item drift silently mis-routes menu selections. Mitigation: the both-platform ordinal table test (step 3) is the guard — non-negotiable.
- Risk: menu snapshot tests (prior plan) break on the new item. Mitigation: update snapshots intentionally; review diff.
- Risk: reimplementing confirm/delete duplicates `appuninstall` logic. Mitigation: accepted per codebase precedent (appuninstall already does this); keep the cleanup model thin; do NOT refactor/extract the monolithic uninstall `App` (respects the plan Non-Goal).
- Risk: synchronous aggregation freezes the TUI. Mitigation: async goroutine+channel load (step 5) + per-detector timeout (step 2).

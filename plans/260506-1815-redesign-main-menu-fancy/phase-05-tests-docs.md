---
phase: 5
title: "Tests + README screenshot update"
status: completed
priority: P2
effort: "30m"
dependencies: [4]
---

# Phase 5: Tests + Docs

## Overview

Lock in the new menu render via snapshot tests (mirrors existing `internal/tui/*_test.go` pattern), and refresh the README screenshot + add `--fancy` to the CLI doc table.

## Requirements

- Functional: snapshot test for narrow + wide layouts, with cursor at index 0 and 2.
- Functional: `--fancy` documented in README CLI flags section.

## Related Code Files

- Create: `internal/tui/menu/menu_test.go`
- Modify: `README.md`

## Implementation Steps

### Task 5.1 — Snapshot tests

1. Mirror existing TUI test style (check `internal/tui/*_test.go` for pattern).
2. Cases:
   - `TestMenuRender_Wide_CursorTop` — width 120, height 30, cursor 0 → strings.Contains active marker on item 0.
   - `TestMenuRender_Narrow_FallbackSingleColumn` — width 60 → no `JoinHorizontal` artifacts (assert banner appears above first card text).
   - `TestMenuRender_Wide_CursorMid` — cursor 1 → marker on Analyze.
3. Use `strings.Contains` assertions, not byte-level snapshots (terminal size variance).

**Commit:** `test(tui/menu): snapshot tests for layout and cursor states`
**Push:** `git push origin main`

### Task 5.2 — README update

1. Capture new screenshot via CleanShot, save to `docs/img/menu.png`.
2. Update README:
   - Replace existing menu screenshot reference.
   - Add `--fancy` row to the CLI flags table with note "enables animated UI effects (slight CPU cost)".
3. If a `docs/img/` dir doesn't exist, create it.

**Commit:** `docs(readme): update menu screenshot and document --fancy flag`
**Push:** `git push origin main`

## Success Criteria

- [x] `go test ./internal/tui/menu/...` passes.
- [x] README shows new screenshot.
- [x] `--fancy` listed in flags table.
- [x] 2 commits pushed.
- [x] Plan can be archived (`/ck:plan archive`).

## Risk Assessment

- **Risk:** Snapshot tests flaky on color profile differences in CI. **Mitigation:** assert structural strings only; never raw ANSI.
- **Risk:** README screenshot path drift. **Mitigation:** keep relative path `docs/img/menu.png`.

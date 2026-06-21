---
phase: 4
title: "JSON output and docs"
status: completed
priority: P2
effort: "S"
dependencies: [3]
---

# Phase 4: JSON output and docs

## Overview

Add scriptable `--json` output for the cleanup advisor (consistent with `analyze --json`) and document the feature in the README. Closes the plan.

## Requirements

- Functional: a non-TTY / `--json` invocation of the cleanup path emits stable JSON (name, manager, version, role, installedAt, size) instead of launching the TUI.
- Non-functional: stable field names; omitempty for zero `installedAt`; matches the existing `analyze --json` convention.

## Architecture

- **A `cleanup` cobra subcommand is required** — there is currently no `newCleanupCmd`; `main.go:74-76` registers `update`/`purge`/`analyze` via `root.AddCommand`. Add `root.AddCommand(newCleanupCmd())` so `omniclean cleanup [--json] [--manager ...] [--dry-run]` exists (the menu→`SelectCleanup` path from Phase 3 covers interactive launch; this subcommand covers scripting).
- Reuse the real `analyze` JSON convention: `analyze` decides via `shouldEmitJSON` using `os.Stdout.Stat() & os.ModeCharDevice` (`cmd/omniclean/analyze.go:71-80`) plus an explicit `--json` flag — **no isatty/3rd-party dep**. Mirror exactly: emit JSON when `--json` set OR stdout is not a char device; else launch the TUI.
- JSON shape:

  ```json
  [{"name":"foo","manager":"brew","version":"1.2","role":"orphan","installedAt":"2025-01-04T00:00:00Z","size":1048576}]
  ```

  `installedAt` omitted when zero. `role` ∈ {manual, orphan} (only removable candidates emitted, matching the TUI filter).

## Related Code Files

- Create: `cmd/omniclean/cleanup.go` (`newCleanupCmd` + `--json`/`shouldEmitJSON` branch), `cmd/omniclean/cleanup_json_test.go`
- Modify: `cmd/omniclean/main.go` (`root.AddCommand(newCleanupCmd())`), README.md
- Reference: `cmd/omniclean/analyze.go:71-80` (`shouldEmitJSON`), `internal/analyze/json.go` (struct-tag + omitempty convention)

## Implementation Steps (TDD)

1. **Test first** — feed aggregated fixture packages to the JSON encoder; assert exact field names, `role` lowercase, `installedAt` omitted when zero.
2. Implement `newCleanupCmd` + JSON branch reusing `shouldEmitJSON`; register in `main.go`; green.
3. Update README: new "Cleanup suggestions" feature bullet + capability-tier table (Full: brew/apt[/flatpak if verified] · leaf-only: cargo/pip/npm/snap/scoop · Unknown: winget/choco) + `cleanup --json` example. `InstalledAt` described with the "best-effort, not true install date" caveat.
4. `go test ./...` green; `gofmt`/`go vet` clean.

## Success Criteria

- [x] `cleanup --json` (or non-TTY) emits stable JSON with `role`/`installedAt`; verified by test.
- [x] `installedAt` omitempty behavior correct.
- [x] README documents the feature, manager capability tiers, and JSON usage.
- [x] Full `go test ./...` green; `go vet ./...` clean; no new deps.

## Risk Assessment

- Risk: JSON field drift vs `analyze --json` style. Mitigation: mirror the existing struct-tag convention exactly; one test pins the shape.
- Risk: README overstates support for leaf-only/Unknown managers. Mitigation: capability-tier table states honestly which managers yield Orphan suggestions.

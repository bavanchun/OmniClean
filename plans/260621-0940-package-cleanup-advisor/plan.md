---
title: "Package Cleanup Advisor (role-based removable classification)"
description: "Reframe of 'detect package usage time' into a role-based removable-package advisor. Adds Role + InstalledAt to pkg.Package, an optional RemovableClassifier capability interface (brew/apt/flatpak), and a new Cleanup-suggestions TUI mode. No usage-time heuristics. TDD per phase."
status: completed
priority: P2
branch: "main"
tags: [tui, detector, cleanup, tdd]
blockedBy: []
blocks: []
created: "2026-06-21T02:44:18.315Z"
createdBy: "ck:plan"
source: skill
---

# Package Cleanup Advisor (role-based removable classification)

## Overview

Replace the infeasible "package usage time" idea with a **role-based cleanup advisor**. Each package is classified by the manager's own bookkeeping into `Manual` (top-level leaf), `Orphan` (auto dep no longer required — safest to remove), `Dependency` (protect/hide), or `Unknown` (manager can't tell). Surfaced in a new "Cleanup suggestions" TUI mode that reuses the existing confirm/leftover/delete flow. No atime / shell-history / Spotlight heuristics — trustworthy-or-silent.

Source brainstorm: `plans/reports/brainstorm-260621-0940-package-cleanup-advisor-report.md`.

**Architecture decision (approved):** optional capability interface `RemovableClassifier`; detectors opt in via type assertion. brew/apt/flatpak implement it; the other 7 degrade to `Unknown` with zero churn (KISS + YAGNI + DRY).

**Per-manager capability tiers:**
- Full (Manual + Orphan): brew, apt
- Leaf-only (Manual; all top-level): flatpak, cargo, pip, npm, snap, scoop
  - **flatpak resolved (Phase 2): leaf-only.** No documented read-only/dry-run form of `flatpak uninstall --unused` exists; per trustworthy-or-silent we never run a mutating command to detect orphans, so flatpak yields Manual only.
- Unknown (no dep graph): winget, choco

**Red-team reconciled (deep-mode gate).** Verified against real code; corrections folded into phases:
- Menu `Selection` is computed **positionally** (`menu.go:208` `Selection(cursor+1)`) — Phase 3 pins exact enum ordinal + a both-platform mapping test; "before SelectQuit" was a false anchor.
- The uninstall confirm/delete flow lives as **unexported symbols inside a monolithic `App`** — cannot be imported. Phase 3 **reimplements** a self-contained cleanup TUI following the shipped `internal/tui/appuninstall/` precedent (which also reimplements, not reuses). This respects the Non-Goal below.
- Existing `CommandRunner` fakes return one canned string; Phase 2 adds an **args-switching, call-recording fake** for multi-command Classify.
- Cleanup aggregation runs **async + per-detector context timeout**; apt delete needs **sudo sequencing** (Phase 3).
- A **`cleanup` cobra subcommand** must be registered (Phase 4) — none exists today.

## Approach: TDD

Each phase is tests-first. Tests use the existing injected `CommandRunner` fake pattern (`internal/detector/brew_test.go`) — no real package-manager invocation in unit tests. Write the failing test, then the implementation, then green.

## Success Criteria (plan-level)

- [x] Cleanup mode lists only `Manual`/`Orphan`; never shows `Dependency` as removable.
- [x] brew/apt/flatpak classifiers verified against fake-runner fixtures.
- [x] `go test ./...` green; existing uninstall flow unchanged (no regression).
- [x] `--json` emits stable `role` + `installedAt` fields.
- [x] No new third-party dependencies in `go.mod`.

## Non-Goals

- No usage-time / last-used timestamp of any kind.
- No new dependency-graph implementation for winget/choco.
- **Don't modify or extract the existing uninstall list TUI / `App`.** The cleanup TUI is a separate self-contained package (mirrors `appuninstall` precedent); it does not refactor, export, or alter the shipped uninstall flow.

## Phases

| Phase | Name | Status |
|-------|------|--------|
| 1 | [Model and capability interface](./phase-01-model-and-capability-interface.md) | Completed |
| 2 | [Classifiers for brew/apt/flatpak](./phase-02-classifiers-for-brew-apt-flatpak.md) | Completed |
| 3 | [Cleanup aggregation and TUI mode](./phase-03-cleanup-aggregation-and-tui-mode.md) | Completed |
| 4 | [JSON output and docs](./phase-04-json-output-and-docs.md) | Completed |

## Dependencies

No cross-plan dependencies. The two prior plans (`260506-1815-redesign-main-menu-fancy`, `260507-0054-fix-tui-and-add-app-uninstall`) are both `completed`; this plan reuses their shipped menu + uninstall-flow components but does not block on them.

## Open Questions

1. ~~Reuse vs reimplement uninstall TUI~~ — RESOLVED: reimplement self-contained (codebase `appuninstall` precedent + Non-Goal). 
2. ~~**flatpak orphan command** — is there a reliably read-only form?~~ — RESOLVED (Phase 2): no read-only/dry-run form of `flatpak uninstall --unused`; flatpak ships **leaf-only** (Manual only). No mutating command used to detect.
3. apt `apt-get autoremove --dry-run` "Remv" line format across Debian/Ubuntu/apt 2.x — confirm distro matrix before Phase 2 impl.
4. Menu label default "Cleanup Suggestions" (vs "Stale packages") — cosmetic, Phase 3.

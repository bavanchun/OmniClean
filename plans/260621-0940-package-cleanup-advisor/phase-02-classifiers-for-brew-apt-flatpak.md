---
phase: 2
title: "Classifiers for brew/apt/flatpak"
status: completed
priority: P1
effort: "M"
dependencies: [1]
---

# Phase 2: Classifiers for brew/apt/flatpak

## Overview

Implement `RemovableClassifier` on the three detectors with real dependency bookkeeping: brew, apt, flatpak. Each gets Manual/Orphan classification plus best-effort `InstalledAt`. Tests-first with the injected `CommandRunner` fake — no real package-manager calls.

## Requirements

- Functional: `brew.Classify`, `apt.Classify`, `flatpak.Classify` mark each package `Manual`/`Orphan`/`Dependency` from read-only manager queries.
- Non-functional: all commands read-only (`--dry-run`/list/query); respect `Available()`; no sudo escalation in classify path; deterministic, fakeable.

## Architecture

Per-manager signal sources (read-only):

| Manager | Manual (leaf) | Orphan (removable dep) | InstalledAt (best-effort) |
|---------|---------------|------------------------|---------------------------|
| brew | `brew leaves` | `brew autoremove -n` (dry-run) | Cellar dir mtime (`{prefix}/Cellar/{name}` via `brew --prefix`) |
| apt | `apt-mark showmanual` | `apt-get autoremove --dry-run` (parse "Remv {pkg}" lines) | `/var/lib/dpkg/info/{name}.list` mtime |
| flatpak | `flatpak list --app --columns=application` | **VERIFY read-only form first** (see risk) | install dir mtime (optional) |

Each detector wraps its queries in a **bounded `context.WithTimeout`**; on timeout/error the classifier returns packages as `RoleUnknown` rather than hanging or guessing.

Classification logic per detector:
1. Build a set of manual names, a set of orphan names.
2. For each package: in orphan set → `RoleOrphan`; else in manual set → `RoleManual`; else → `RoleDependency`.
3. Annotate `InstalledAt` when the cheap source resolves; leave zero otherwise. (Note: dpkg `.list` / Cellar mtime track last file-list write / relink, NOT true install time — best-effort only, never decision-driving; README must caveat, not call it "install date" unqualified.)

**No usage-time.** If a manager query is unavailable at runtime, return packages with `RoleUnknown` rather than guessing.

### Test harness note (REQUIRED — existing fake is insufficient)

The current `CommandRunner` fakes (`brew_test.go:62`, `apt_test.go:77`) ignore name+args and return ONE canned string. `Classify` issues TWO distinct commands (e.g. `brew leaves` AND `brew autoremove -n`) needing DIFFERENT outputs to assert Manual vs Orphan vs Dependency. Phase 2 MUST introduce an **args-switching, call-recording fake**:
- dispatch output on `args` (e.g. match `"leaves"` vs `"autoremove"`).
- record each `(name, args)` call so the success criterion "only read-only commands issued" is assertable.
This is a legitimate new test scaffold, not "the existing pattern."

## Related Code Files

- Modify: `internal/detector/brew.go`, `internal/detector/apt.go`, `internal/detector/flatpak.go` (add `Classify`)
- Test: `internal/detector/brew_test.go`, `apt_test.go`, `flatpak_test.go` (extend with classify cases + fake runner fixtures)
- Reference (no change): `internal/detector/detector.go` (CommandRunner), `internal/detector/classifier.go` (Phase 1 interface)

## Implementation Steps (TDD)

0. **Build the args-switching, call-recording fake** (shared test helper in `detector` test package) — dispatches canned output per `args`, records all calls.
1. **brew test first** — fake dispatches `leaves` + `autoremove` outputs; assert leaf → `RoleManual`, autoremove target → `RoleOrphan`, linked-not-leaf → `RoleDependency`. Assert only read-only commands recorded.
2. Implement `brew.Classify` (bounded context timeout); green.
3. **apt test first** — fake dispatches `apt-mark showmanual` + `apt-get autoremove --dry-run`; assert roles + read-only. Implement; green.
4. **flatpak**: FIRST pin the read-only orphan command (risk below). If a reliable read-only orphan form exists → test+implement Manual+Orphan. If NOT → implement **leaf-only** (Manual only) and demote flatpak to leaf-only tier in `plan.md` + README.
5. Add one `InstalledAt`-present and one `InstalledAt`-absent case per manager (mtime source faked or skipped via interface seam).
6. Add a timeout/unavailable case per manager asserting `RoleUnknown`, no panic.
7. `go test ./internal/detector/...` → green.

## Success Criteria

- [x] Args-switching recording fake exists; brew/apt classifiers verified for all three roles.
- [x] flatpak orphan command verified read-only OR flatpak demoted to leaf-only (tier reconciled in plan.md + README).
- [x] Classify path issues only read-only commands (asserted via recorded call log).
- [x] Each `Classify` bounded by context timeout; unavailable/slow query → `RoleUnknown`, no panic, not fatal to caller.
- [x] `InstalledAt` populated when source present, zero otherwise.
- [x] `go test ./...` green.

## Risk Assessment

- Risk: `apt-get autoremove --dry-run` "Remv {pkg}" format varies Debian vs Ubuntu vs apt 2.x. Mitigation: parse stable "Remv " lines; fixture from known version; parse-miss → `RoleUnknown`, not crash. (Open Q #3: confirm distro matrix.)
- Risk: **flatpak has no documented stable `--dry-run` for `uninstall --unused`** (red-team). Mitigation: derive orphans from `flatpak list --runtime` vs app deps if a read-only form exists; else ship flatpak leaf-only and adjust the tier table — do NOT run a mutating `uninstall` to "detect."
- Risk: `brew autoremove -n` not guaranteed instant (tap/Cellar scan). Mitigation: bounded context timeout → degrade to Unknown; tests use fakes.
- Risk: mtime is a weak `InstalledAt` proxy. Mitigation: best-effort, never drives the removable decision; README caveats wording.

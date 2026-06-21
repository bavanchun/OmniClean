---
phase: 1
title: "Model and capability interface"
status: completed
priority: P1
effort: "S"
dependencies: []
---

# Phase 1: Model and capability interface

## Overview

Extend `pkg.Package` with a `Role` enum and optional `InstalledAt`, and define the opt-in `RemovableClassifier` interface. No behavior change; pure additive groundwork. Follows the existing `Size int64` "zero = unknown" idiom.

## Requirements

- Functional: `pkg.Package` carries a classification role and optional install time; a capability interface lets detectors opt into classification.
- Non-functional: zero churn to the 9 detectors that don't classify; no new deps; backward-compatible (existing struct literals still compile — new fields are additive and zero-valued).

## Architecture

- New type `pkg.Role` (string enum): `RoleUnknown` (zero value), `RoleManual`, `RoleOrphan`, `RoleDependency`. Zero value MUST be `RoleUnknown` so unclassified packages are safe by default.
- Add to `pkg.Package`: `Role Role` and `InstalledAt time.Time` (zero = unknown, render `—`).
- New interface in `internal/detector` (capability, NOT added to `Detector`):

  ```go
  // RemovableClassifier is an optional capability. Detectors that can consult
  // their manager's dependency bookkeeping implement it; callers use a type
  // assertion (d.(RemovableClassifier)) and skip detectors that don't.
  type RemovableClassifier interface {
      // Classify annotates packages in place (or returns a copy) with Role and,
      // when cheaply available, InstalledAt. Read-only: no uninstall side effects.
      Classify(ctx context.Context, pkgs []pkg.Package) ([]pkg.Package, error)
  }
  ```

## Related Code Files

- Modify: `internal/pkg/package.go` (add `Role` type + `Package.Role`, `Package.InstalledAt`)
- Create: `internal/detector/classifier.go` (the `RemovableClassifier` interface + a small `Classify` helper that no-ops for non-classifiers)
- Test: `internal/pkg/package_test.go` (role zero-value + render), `internal/detector/classifier_test.go`

## Implementation Steps (TDD)

1. **Test first** — `package_test.go`: assert `pkg.Role("")` zero value equals `RoleUnknown`; assert a `Package` with `Role`/`InstalledAt` unset behaves like today (Desc unchanged).
2. **Test first** — `classifier_test.go`: a fake detector NOT implementing `RemovableClassifier` passed to the helper returns packages unchanged with `RoleUnknown`; a fake that DOES implement it has `Classify` invoked.
3. Implement `Role` enum + fields in `package.go`.
4. Implement `classifier.go`: interface + `ClassifyIfSupported(ctx, d Detector, pkgs)` helper using type assertion.
5. Run `go test ./internal/pkg/... ./internal/detector/...` → green.

## Success Criteria

- [x] `pkg.Role` zero value is `RoleUnknown`; `Package` gains `Role` + `InstalledAt` (additive).
- [x] `RemovableClassifier` defined in `internal/detector`; NOT part of the core `Detector` interface.
- [x] Helper degrades non-classifier detectors to `RoleUnknown` without error.
- [x] New + existing tests green; `go build ./...` clean.
- [x] No new entries in `go.mod`.

## Risk Assessment

- Risk: adding fields breaks positional struct literals. Mitigation: codebase uses keyed literals (`pkg.Package{Name: ...}`); grep-verify no positional `pkg.Package{` literals before merge.
- Risk: `time` import bloats `pkg`. Mitigation: stdlib only, already idiomatic.

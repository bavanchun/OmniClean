# TDD Discipline

This repo locks behavior with tests. Write tests first, then make them pass.

## Per phase

1. Write the failing test(s) for the phase's behavior first (red).
2. Implement the minimum to pass (green).
3. Refactor with tests green. Never weaken or delete a test to make a change pass — fix the cause.

## Detector / classifier tests

- Use the injected `CommandRunner` fake and the args-switching recording fake in
  `internal/detector/classify_fake_test.go`. Unit tests are hermetic: no network, no real package
  manager at test time.
- Lock parsers against **real captured distro output** stored under `internal/detector/testdata/`,
  with provenance (capture command + distro version) recorded in `testdata/README.md`. If a real
  capture is impractical, hand-author a faithful sample and **label it synthetic** — never pass
  synthetic output off as real.
- Naming follows the existing convention: `Test<Mgr>_ListPackages`, `Test<Mgr>_Classify`,
  `Test<Mgr>_Classify_RealFixtures`, `Test<Mgr>_Classify_UnavailableDegradesToUnknown`,
  `Test<Mgr>_Metadata`. Assert read-only behavior (no sudo, dry-run present) by inspecting the
  fake's recorded calls.

## Trustworthy-or-silent

A classifier must be correct or silent. If a manager has no reliable native orphan query, ship it
leaf-only (every package → `RoleManual`, like flatpak) rather than guessing. Degrade to
`RoleUnknown` on failure. A wrong "removable" verdict is worse than no verdict.

See also: [go-conventions](go-conventions.md), [git-workflow](git-workflow.md).

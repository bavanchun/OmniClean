# Go Conventions & Quality Gates

OmniClean is Go 1.25 (`github.com/bavanchun/OmniClean`). Match the existing code; don't reinvent.

## Quality gates (must pass before a phase is done)

- `gofmt -l .` prints nothing (everything formatted).
- `go vet ./...` is clean.
- `go build ./...` succeeds.
- `go test ./... -race` is green (the CI matrix runs 3 OSes with `-race`).
- **No new third-party dependencies** without explicit human approval. Prefer the stdlib.

## Package-manager detector pattern

- Each manager is one file in `internal/detector/` implementing the `Detector` interface
  (`internal/detector/detector.go`). Optional removable classification implements
  `RemovableClassifier` (`internal/detector/classifier.go`) — picked up automatically by
  `ClassifyIfSupported` via type assertion; no registry change for the capability itself.
- Register a new detector in `AllDetectors()` (`internal/detector/registry.go`) and add its
  `ManagerType` constant in `internal/pkg/package.go`. The `ManagerType` value MUST equal the
  detector's `Name()` (the CLI `--manager` flag matches on it).
- Reuse the shared helpers in `classifier.go`: `lineSet`, `markAllUnknown`, `classifyTimeout`,
  and the `statFunc` mtime seam. Don't duplicate them.
- Use the package-local **`LookPath("name")`** seam for binary detection, not `exec.LookPath`.
- Command output is parsed under a pinned **`LC_ALL=C` / `LANG=C`** locale (`commandEnv()` in
  `detector.go`). Build parsers around stable English tokens; never depend on the host locale.
- `Classify` is read-only: no sudo, no mutation; on probe failure degrade to `RoleUnknown`
  (safe-by-default), never guess a package is removable.

## Style

- Go files use lowercase package-file names and `snake_case` where a separator is needed; types
  use `PascalCase`, unexported identifiers `camelCase` — follow gofmt + the surrounding code.
- Consider modularizing a file once it passes ~200 LOC, if it has a clean logical boundary.
- Write descriptive comments for non-obvious behavior; match the existing comment density.

See also: [tdd](tdd.md), [00-read-first](00-read-first.md).

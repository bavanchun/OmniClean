# CI/CD Professional Overhaul — Modularized Workflows, Supply-Chain Hardening, and Signed Releases

**Date**: 2026-06-25 14:42
**Severity**: Medium
**Component**: CI/CD (GitHub Workflows, GoReleaser, golangci-lint, Makefile, templates, CODEOWNERS)
**Status**: Resolved

## What Happened

Completed 8-phase implementation utilizing file-disjoint parallel agents:
1. Migrated `.golangci.yml` to the v2 schema config (version 2) compatible with `v2.12.2`.
2. Created `.github/workflows/lint.yml` containing linting (`golangci-lint@v2.12.2`), tidy check, and actionlint.
3. Created `.github/workflows/test.yml` with a 3-OS matrix (`go test -race` + coverage), Codecov upload, and ported the real-apt smoke-test (`classify-smoke`).
4. Created `.github/workflows/build.yml` compiling binaries on 3 OSes and uploading artifacts.
5. Created `.github/workflows/security.yml` (govulncheck + dependency review) and `codeql.yml` (CodeQL scanning).
6. Hardened GoReleaser and release pipeline: modified `.goreleaser.yml` to support SPDX-JSON SBOMs (syft), keyless signing (cosign OIDC), nfpm packaging (deb/rpm), and Homebrew tap.
7. Fixed the release Go version bug by leveraging `go-version-file: go.mod` across all workflows.
8. Created `.github/workflows/release-validate.yml` to run snapshot checks on pull requests.
9. Added governance elements: `.github/CODEOWNERS`, pull request template, bug/feature issue templates, and modernized `Makefile` development targets.
10. Created the public repository `bavanchun/homebrew-tap` and set up repository secrets.
11. Merged all 6 disjoint PRs and performed the final cutover by removing the legacy monolithic `ci.yml` file.

## Technical Details

- **Workflow Splitting**:
  - Replaced monolithic `ci.yml` with concern-specific workflows (`lint.yml`, `test.yml`, `build.yml`, `security.yml`, `codeql.yml`, `release-validate.yml`, `release.yml`).
  - Added least-privilege permissions and auto-cancel concurrency groups to every workflow file.
- **SHA Action Pinning**:
  - All actions pinned to full 40-character commit SHAs (e.g. `actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683`) with version comments.
- **GoReleaser (v2.16)**:
  - Configured SBOMs, signing, nfpm, and homebrew_casks blocks using v2-compliant keys (e.g. `homebrew_casks:` and `formats:` instead of deprecated keys).
  - Snapshot validation uses `--skip=publish,sign` since signing requires OIDC on a real tag.
- **Makefile DX**:
  - `vet`, `vuln` (govulncheck), `tidy-check`, `sbom`, `release-snapshot` and `lint` targets introduced to mirror CI environments locally.
  - `install-tools` updated to pin golangci-lint to `v2.12.2` and goreleaser to `v2.16.0`.

## Lessons Learned

1. **Workspace and File Disjointness for Agent Parallelism**: Splitting a large configuration file into separate, independent workflows allowed multiple subagents to work concurrently on different branches without git conflicts.
2. **Worktree Conflicts**: Git worktrees created by subagents prevent branch deletion when local branches are active. Removing the worktrees explicitly (`git worktree remove`) resolves the conflict.
3. **Cosign OIDC keyless constraints**: Cosign keyless blob signing requires OIDC authentication, which only runs under GHA environments. Snapshot runs must skip the `sign` step to avoid credential failures.

## Next Steps

- Set real tokens for `CODECOV_TOKEN` and `HOMEBREW_TAP_GITHUB_TOKEN` on the repository.
- Verify the first real tagged release `v*` compiles, signs, packages, and pushes to `bavanchun/homebrew-tap`.

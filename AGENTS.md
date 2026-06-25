# AGENTS.md

Canonical instructions for **every** AI agent working in this repository. This is the single
front-door — Claude (via `CLAUDE.md` → `@AGENTS.md`), Antigravity, Codex, and any other harness
all read this file.

## Mandate

**You MUST read every file in [`rules/`](rules/) before writing or modifying any code.**
The rules are the project's source of truth for how to work here. They override your defaults
where they conflict; explicit human instructions override the rules.

## Rules index

| Rule | Purpose |
|------|---------|
| [`rules/00-read-first.md`](rules/00-read-first.md) | What context to load first; KISS·YAGNI·DRY; plan & doc conventions |
| [`rules/go-conventions.md`](rules/go-conventions.md) | Go 1.25 quality gates; detector/classifier pattern; `LC_ALL=C` parsing |
| [`rules/tdd.md`](rules/tdd.md) | Tests-first; hermetic fakes; real-distro fixtures; trustworthy-or-silent |
| [`rules/git-workflow.md`](rules/git-workflow.md) | No code on `main`; branch/commit/PR-per-phase; conventional commits; secrets |

## Priority harnesses

Claude, Antigravity, and Codex are the primary harnesses; all of them load this file and the
`rules/` it points to.

- **Claude Code:** `CLAUDE.md` imports this file with `@AGENTS.md`, which links the `rules/`.
- **Codex / Antigravity:** read this `AGENTS.md` (agents.md convention) and the linked `rules/`.

Harness-specific shortcuts (e.g. the Claude-only `/vchun-git prc`) are noted in the relevant rule
as shortcuts; the generic intent always applies to every agent.

## Project at a glance

OmniClean — a cross-platform Go (1.25) TUI/CLI that unifies package-manager uninstall, leftover
detection, project-artifact purge, disk analysis, and cleanup suggestions. See `README.md` and
`docs/codebase-summary.md` for detail.

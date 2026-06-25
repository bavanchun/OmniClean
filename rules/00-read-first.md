# 00 — Read First

Read these before writing or modifying any code in this repo.

## Context to load

1. `README.md` — what OmniClean is, the subcommands, the supported managers.
2. `docs/codebase-summary.md` — architecture and module map.
3. The active plan, if any: `plans/<plan-dir>/plan.md` and every `plans/<plan-dir>/phase-*.md`.
   Implement only what the current phase specifies.

> Note: `plans/` is git-ignored (local working artifacts). If you don't see it in a fresh clone,
> ask the human for the plan path or the task scope.

## Engineering principles

- **KISS · YAGNI · DRY**, in that order. Build what the request needs, not what it might need.
- Implement real behavior. No fake data, mocks, or shortcuts left in to satisfy a check.
- Keep changes scoped to the request and the contracts it touches.
- Prefer changing existing files / local helpers over inventing new abstractions.

## Plan & doc conventions

- Plans use checkbox success criteria: `- [ ]` (open) / `- [x]` (done). Progress is counted from
  these — keep them accurate; never mark a box done before its criterion is verified.
- Phase files carry frontmatter (`phase`, `status`, `priority`, `dependencies`) and the standard
  sections (Overview, Requirements, Architecture, Related Code Files, Implementation Steps,
  Success Criteria, Risk Assessment). Match the existing house style.
- Keep important docs in `docs/`. Update them only when user-visible behavior, setup, commands,
  architecture, or public contracts change.

See also: [go-conventions](go-conventions.md), [tdd](tdd.md), [git-workflow](git-workflow.md).

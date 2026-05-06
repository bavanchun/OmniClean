---
title: "Redesign main menu — Path B (full xin xò + animation)"
status: in-progress
priority: P2
created: 2026-05-06
scope: project
owner: bavanchun
sourceReport: reports/260506-1809-redesign-main-menu.md
blockedBy: []
blocks: []
---

# Plan: Redesign Main Menu (Path B — Full Animated)

## Goal

Replace the current flat vertical list main menu with a **2-column dashboard** (brand panel + action cards) using Charm v2 capabilities: gradient borders, ASCII banner, adaptive color, animated spotlight via `--fancy` flag.

Source analysis: `reports/260506-1809-redesign-main-menu.md`.

## Non-Goals

- Don't touch downstream TUIs (uninstall / analyze / purge views).
- Don't add new menu entries.
- Don't replace the theme tokens system; only add semantic aliases.

## Constraints

- Stack: `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2` (already in `go.mod`).
- New deps: **none** (banner hard-coded ASCII).
- Fallback to single column when `width < 80`.
- Animation OFF by default; opt-in via `--fancy`.
- All text must work on Windows Terminal + macOS iTerm2; legacy `cmd.exe` may degrade to solid borders.

## Phases

| # | Phase | Status | Effort | Commits expected |
|---|-------|--------|--------|------------------|
| 1 | Refactor base (no visual change) | pending | 30m | 1 |
| 2 | 2-column layout + cards + ASCII banner | pending | 90m | 2-3 |
| 3 | Polish: bubbles/help + adaptive color + number shortcuts | pending | 45m | 1-2 |
| 4 | Animation + `--fancy` flag | pending | 60m | 2 |
| 5 | Tests + README screenshot update | pending | 30m | 1 |

Total: ~4h, 7-9 commits.

## Commit & Push Protocol (MANDATORY)

After **each completed task** within a phase:

1. Stage only the files touched by that task (no `git add -A`).
2. Commit with **Conventional Commits** spec (`https://www.conventionalcommits.org/en/v1.0.0/`).
   - Format: `<type>(<scope>): <subject>` + optional body.
   - Types used here: `feat`, `refactor`, `style`, `test`, `docs`, `chore`.
   - Scope: `tui/menu` (or `theme` when touching tokens).
   - **DO NOT** include `Co-authored-by: Cursor` trailer (use `git commit-tree` plumbing if a hook injects it; see `reports/260506-1809-redesign-main-menu.md` precedent).
3. **Push immediately** after commit: `git push origin main`.
   - `origin` is configured with multi push URL → pushes to GitHub + GitLab in one call.
4. Verify with `git status` (clean) and `git log -1 --format=%B` (no unwanted trailer).

## Success Criteria

- [ ] Menu renders with 2-column dashboard at width ≥ 80.
- [ ] Active card uses gradient border (Lipgloss v2 blend).
- [ ] Number shortcuts `1/2/3` jump-select.
- [ ] `?` toggles full help, footer auto-resizes.
- [ ] Single-column fallback below width 80.
- [ ] `--fancy` enables animated banner spinner + gradient offset rotation.
- [ ] No new dependencies in `go.mod` (banner is hard-coded).
- [ ] All commits follow Conventional Commits, no `Co-authored-by` trailer.
- [ ] Each commit pushed to `origin` (GitHub + GitLab) before next task starts.
- [ ] Existing `go test ./...` still green.

## Risks

| Risk | Mitigation |
|------|------------|
| `BorderForegroundBlend` API name drift on lipgloss v2 minor | Probe API in Phase 2 first, fall back to solid `BorderForeground(Primary)` if missing |
| Animation tick CPU cost when `--fancy` on | Keep tick interval ≥ 150ms; document trade-off in README |
| Width < 80 vỡ layout | Phase 2 step 6 implements explicit fallback path with snapshot test |
| Hook re-injects `Co-authored-by` | Use `git commit-tree` plumbing as in last 2 commits |

## Files

### Create
- `internal/tui/menu/banner.go` — hard-coded ASCII banner (Phase 2)
- `internal/tui/menu/styles.go` — package-level styles (Phase 1)
- `internal/tui/menu/keys.go` — `bubbles/key` + `bubbles/help` integration (Phase 3)
- `internal/tui/menu/menu_test.go` — snapshot tests (Phase 5)

### Modify
- `internal/tui/menu/menu.go` — main model + view logic
- `cmd/omniclean/main.go` — add `--fancy` flag (Phase 4)
- `internal/tui/theme/tokens.go` — add semantic aliases only (Phase 2)
- `README.md` — updated screenshot + flag docs (Phase 5)

### Delete
- (none)

## Out-of-Scope (explicitly)

- Replacing `bubbles/list` — overkill for 3 items.
- Adding logo via `x/mosaic` (image render). Cool but YAGNI.
- Live system status (managers count) — punt to follow-up plan.

## References

- Report: `reports/260506-1809-redesign-main-menu.md`
- Lipgloss v2 upgrade: https://github.com/charmbracelet/lipgloss/blob/main/UPGRADE_GUIDE_V2.md
- Border blend PR: https://github.com/charmbracelet/lipgloss/pull/560
- Bubbles v2: https://github.com/charmbracelet/bubbles/releases/tag/v2.1.0
- Tabs example: https://github.com/charmbracelet/bubbletea/blob/main/examples/tabs/main.go

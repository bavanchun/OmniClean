---
phase: 2
title: "2-column layout + cards + ASCII banner"
status: completed
priority: P1
effort: "90m"
dependencies: [1]
---

# Phase 2: Layout, Cards, Banner

## Overview

Replace the vertical list with a 2-column dashboard: brand panel left (ASCII banner + meta) + action cards right with gradient border on the active card.

## Requirements

- Functional: at width ≥ 80, render brand panel (≤ 30 cols) + cards (≥ 45 cols). Below 80, fall back to stacked single column.
- Non-functional: card render allocates O(items) strings per frame; no per-char loops.

## Architecture

```
┌──────────────┐  ┌─────────────────────┐
│ BANNER       │  │ ▌ [1] Uninstall ... │
│ ✦ OmniClean  │  │     Search & remove │
│ vX.Y.Z       │  └─────────────────────┘
│              │  ┌─────────────────────┐
│              │  │ ▌ [2] Analyze ...   │
└──────────────┘  └─────────────────────┘
                  ┌─────────────────────┐
                  │ ▌ [3] Purge ...     │
                  └─────────────────────┘
```

- `JoinVertical(Left, cards...)` for the right column.
- `JoinHorizontal(Top, brandPanel, "   ", cardsCol)` for body.
- Active card: `BorderForegroundBlend(Primary, Accent)`. If API absent on installed version, fallback `BorderForeground(Primary)` (Phase 2.1 probes).
- Inactive card: `BorderForeground(SurfaceBorder)`.
- Quit removed from `menuItems`; handled exclusively via `q` key.

## Related Code Files

- Create: `internal/tui/menu/banner.go`
- Modify: `internal/tui/menu/styles.go`, `internal/tui/menu/menu.go`, `internal/tui/theme/tokens.go`

## Implementation Steps

### Task 2.1 — Probe Lipgloss v2 blend API + add semantic tokens

1. Write a tiny throwaway file `/tmp/blend_probe.go` calling `lipgloss.NewStyle().BorderForegroundBlend(...)` to confirm the method exists on the installed version. Delete after.
2. If present → record `useBlend = true`. If not → `useBlend = false`, document fallback.
3. Add to `internal/tui/theme/tokens.go`:
   ```go
   const (
       BorderActive   = Primary
       BorderActive2  = Accent
       BorderIdle     = SurfaceBorder
       BarActive      = Primary
       BrandPanelBg   = SurfaceElevated
   )
   ```
4. `go build ./...`.

**Commit:** `feat(theme): add semantic border/bar/panel tokens for menu`
**Push:** `git push origin main`

### Task 2.2 — ASCII banner

1. Create `internal/tui/menu/banner.go`:
   ```go
   package menu

   // bannerLines is a hand-tuned 5-row ASCII banner for "OMNI".
   // Keeping it short keeps the brand panel under 30 cols.
   var bannerLines = []string{
       " ██████  ███▄ ▄███▓ ███▄    █  ██▓",
       "▒██    ▒ ▓██▒▀█▀ ██▒ ██ ▀█   █ ▓██▒",
       "░ ▓██▄   ▓██    ▓██░▓██  ▀█ ██▒▒██▒",
       "  ▒   ██▒▒██    ▒██ ▓██▒  ▐▌██▒░██░",
       "▒██████▒▒▒██▒   ░██▒▒██░   ▓██░░██░",
   }
   ```
2. Add `renderBrandPanel()` in `menu.go` that joins banner + `✦ OmniClean` + `Unified cleanup toolkit` + version line.
3. Style: panel border rounded, `BorderForeground(BrandPanelBg)`, padding `(1, 2)`.

**Commit:** `feat(tui/menu): add ASCII banner brand panel`
**Push:** `git push origin main`

### Task 2.3 — Card renderer + 2-column layout

1. Add `cardActive`, `cardIdle`, `barStyle` to `styles.go` using semantic tokens.
   - If `useBlend = true`: `cardActive = base.BorderForegroundBlend(BorderActive, BorderActive2)`.
   - Else: `cardActive = base.BorderForeground(BorderActive)`.
2. New `renderCard(i int, item menuItem) string` produces 2-line content (`▌ [n]  Title` + `     desc`) inside the active/idle border style.
3. Drop the `Quit` entry from `menuItems`.
4. Update `Update()` so `q` / `esc` / `ctrl+c` set `SelectQuit`.
5. New `View()` composition:
   ```go
   right := lipgloss.JoinVertical(lipgloss.Left, cards...)
   left  := a.renderBrandPanel()
   body  := lipgloss.JoinHorizontal(lipgloss.Top, left, "   ", right)
   foot  := helpStyle.Render(helpLine)  // bubbles/help comes in Phase 3
   full  := lipgloss.JoinVertical(lipgloss.Center, body, "", foot)
   return tea.NewView(lipgloss.Place(w, h, Center, Center, full))
   ```
6. Manually verify in terminal at widths 120, 100, 90.

**Commit:** `feat(tui/menu): two-column dashboard with action cards`
**Push:** `git push origin main`

### Task 2.4 — Single-column fallback for narrow terminals

1. In `View()`, if `a.width < 80`: skip `JoinHorizontal`, render banner-on-top + cards-below stacked.
2. Verify at width 70.

**Commit:** `feat(tui/menu): single-column fallback when width < 80`
**Push:** `git push origin main`

## Success Criteria

- [ ] Brand panel renders banner + version on the left.
- [ ] 3 cards render on the right at width ≥ 80.
- [ ] Active card visually distinct (gradient if API supports, solid otherwise).
- [ ] Quit removed from list, still triggerable via `q`/`esc`.
- [ ] Width < 80 falls back to single column without overflow.
- [ ] 4 commits pushed.
- [ ] `go build ./...` and `go test ./...` pass.

## Risk Assessment

- **Risk:** `BorderForegroundBlend` not yet released in pinned `lipgloss/v2` version. **Mitigation:** Task 2.1 probe + boolean switch.
- **Risk:** ASCII banner uses Unicode block chars that mis-align in certain fonts. **Mitigation:** Use only chars from `▒░▓█▄▀▐▌` set, all 1-cell wide.
- **Risk:** `JoinHorizontal` height misalignment if banner taller than cards. **Mitigation:** Pad shorter side with empty lines via `lipgloss.NewStyle().Height(n)`.

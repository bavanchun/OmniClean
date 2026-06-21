---
phase: 1
title: "Refactor base (no visual change)"
status: completed
priority: P2
effort: "30m"
dependencies: []
---

# Phase 1: Refactor Base

## Overview

Clean up `menu.go` so subsequent visual work is fast: extract styles, drop hand-rolled centering, drop `max` shim, keep render output identical.

## Requirements

- Functional: visual output **byte-identical** to current menu.
- Non-functional: zero per-frame allocations for styles; cleaner separation of concerns.

## Architecture

- New file `internal/tui/menu/styles.go` holds `var (...)` package-level Lipgloss styles.
- `menu.go` consumes styles via name only.
- `lipgloss.Place(w, h, Center, Center, content)` replaces manual padding loop.
- Built-in `max` (Go 1.21+) replaces local helper.

## Related Code Files

- Create: `internal/tui/menu/styles.go`
- Modify: `internal/tui/menu/menu.go`

## Implementation Steps

### Task 1.1 — Extract styles to package-level

1. Create `internal/tui/menu/styles.go` with:
   ```go
   package menu

   import (
       "charm.land/lipgloss/v2"
       "github.com/bavanchun/OmniClean/internal/tui/theme"
   )

   var (
       titleStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Bold(true)
       subtitleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextSubtle))
       boxStyle           = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(theme.Primary)).Padding(0, 2)
       cursorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary)).Bold(true)
       activeTitleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextStrong)).Bold(true)
       activeDescStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextSubtle))
       inactiveTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextSubtle))
       inactiveDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextMuted))
       helpStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextDim))
   )
   ```
2. Remove the inline `lipgloss.NewStyle()...` chain from `render()`.
3. Run `go build ./...` and visually diff in terminal — output must match.

**Commit:** `refactor(tui/menu): extract styles to package-level vars`
**Push:** `git push origin main`

### Task 1.2 — Replace centering + drop `max` shim

1. Delete the local `func max(a, b int) int`.
2. Replace the manual padding loop at the bottom of `render()` with:
   ```go
   if a.width <= 0 {
       return content
   }
   return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
   ```
3. Run `go build ./...`. Manually launch `go run ./cmd/omniclean` and confirm centering still looks correct.

**Commit:** `refactor(tui/menu): use lipgloss.Place for centering`
**Push:** `git push origin main`

## Success Criteria

- [x] Phase scaffolding committed.
- [x] `styles.go` exists; `menu.go` no longer constructs styles in `render()`.
- [x] `lipgloss.Place` used; no manual `strings.Repeat("\n", ...)` for vertical pad.
- [x] No local `max` function.
- [x] `go build ./...` passes.
- [x] `go test ./...` passes.
- [x] 2 commits pushed (refactor styles, refactor centering).

## Risk Assessment

- **Risk:** `lipgloss.Place` ignores `\n` prefix that current code uses for top-padding. **Mitigation:** strip the leading `\n` from `content` when `width > 0` path is taken.
- **Risk:** Hidden caller depends on local `max`. **Mitigation:** `grep -n "menu.max\|^func max" internal/` before delete.

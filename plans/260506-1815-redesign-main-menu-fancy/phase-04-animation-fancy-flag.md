---
phase: 4
title: "Animation + --fancy flag"
status: completed
priority: P3
effort: "60m"
dependencies: [3]
---

# Phase 4: Animation + `--fancy` Flag

## Overview

Add opt-in liveness: spinner replacing the `✦` star in the brand panel + animated gradient offset on the active card border ("spotlight" effect). Off by default; enable via `--fancy` CLI flag.

## Requirements

- Functional: when `--fancy=false`, the menu is **byte-identical** to Phase 3 output.
- Functional: when `--fancy=true`, banner star animates 4 frames and active card border gradient offset rotates.
- Non-functional: tick interval ≥ 150ms; idle CPU < 2% on M-series Mac.

## Architecture

- New flag `--fancy` (bool) plumbed through `cmd/omniclean/main.go` → `menutui.Run(menutui.Options{Fancy: bool})`.
- `menu.Run` signature breaking-change OK (only 1 caller).
- `App` gains `fancy bool`, `spin spinner.Model`, `tick int`.
- `Init()` returns batch `[spin.Tick, tea.Tick(150ms, ...)]` only when `fancy`.
- Update handles `tea.TickMsg` → increment `a.tick`; rebuild active card style with `BorderForegroundBlendOffset(float64(a.tick%100)/100)`.

## Related Code Files

- Modify: `cmd/omniclean/main.go`, `internal/tui/menu/menu.go`, `internal/tui/menu/styles.go`

## Implementation Steps

### Task 4.1 — `--fancy` flag plumbing

1. In `cmd/omniclean/main.go`:
   ```go
   var fancy bool
   rootCmd.Flags().BoolVar(&fancy, "fancy", false, "enable animated UI effects")
   ```
2. Add type `menutui.Options{Fancy bool}` and change `menutui.Run()` → `menutui.Run(menutui.Options)`.
3. Update single caller in `main.go`.
4. `App.fancy` field set from options.

**Commit:** `feat(cli): add --fancy flag for animated menu`
**Push:** `git push origin main`

### Task 4.2 — Spinner star + ticking gradient offset

1. In `menu.go`:
   ```go
   import "charm.land/bubbles/v2/spinner"

   // in App
   spin spinner.Model
   tick int
   ```
2. `New(opts Options)`:
   ```go
   s := spinner.New()
   s.Spinner = spinner.Spinner{Frames: []string{"✦","✧","✦","✧"}, FPS: 200 * time.Millisecond}
   s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Primary))
   return &App{fancy: opts.Fancy, spin: s, ...}
   ```
3. `Init()`:
   ```go
   if a.fancy {
       return tea.Batch(a.spin.Tick, tickCmd())
   }
   return nil
   ```
   `tickCmd` = `tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })`.
4. `Update()`:
   - `spinner.TickMsg` → forward to `a.spin.Update(msg)`, return cmd.
   - `tickMsg` → `a.tick++`; if `a.fancy` re-issue `tickCmd()`.
5. `renderBrandPanel()` swaps the literal `✦` for `a.spin.View()` when `a.fancy`.
6. `renderCard()` for active card: when `a.fancy` and `useBlend`, call `style.BorderForegroundBlendOffset(float64(a.tick%100)/100.0)`. Else identical to Phase 2.

**Commit:** `feat(tui/menu): animated spinner star and gradient offset under --fancy`
**Push:** `git push origin main`

## Success Criteria

- [ ] Without `--fancy`: render identical to Phase 3 final output.
- [ ] With `--fancy`: star cycles ✦/✧ ~5×/sec; active card border gradient visibly rotates.
- [ ] No tick storms — `tea.Tick` re-issued exactly once per tick.
- [ ] Idle CPU sampled with `top -pid $(pgrep omniclean)` < 2%.
- [ ] 2 commits pushed.

## Risk Assessment

- **Risk:** `BorderForegroundBlendOffset` not exposed → gradient won't animate. **Mitigation:** Phase 2 probe stored as `useBlend`. If false, animation is **only** the spinner.
- **Risk:** Reentrancy bug when spinner + tick both forwarded. **Mitigation:** isolate handlers per msg type; add unit test in Phase 5.
- **Risk:** Terminal flicker on slow SSH. **Mitigation:** `--fancy` is opt-in; document caveat in README.

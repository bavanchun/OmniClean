---
phase: 3
title: "Polish: bubbles/help + number shortcuts + adaptive color"
status: completed
priority: P2
effort: "45m"
dependencies: [2]
---

# Phase 3: Polish

## Overview

Replace hand-rolled help string with `bubbles/help` + `bubbles/key`, add number shortcuts (`1/2/3`) for direct jump, and switch text/border colors to `compat.AdaptiveColor` for light/dark terminals.

## Requirements

- Functional: `?` toggles short ↔ full help; help auto-truncates at narrow widths.
- Functional: `1`, `2`, `3` jump-select corresponding card.
- Non-functional: no breaking change to `Run() (Selection, error)` signature.

## Architecture

- New `internal/tui/menu/keys.go` defines `keyMap` struct implementing `help.KeyMap` interface (`ShortHelp() []key.Binding` + `FullHelp() [][]key.Binding`).
- `App` gains fields `keys keyMap` and `help help.Model`.
- Adaptive color: where the design uses `lipgloss.Color(theme.X)` and X is text/border, swap to `compat.AdaptiveColor{Light, Dark}` derived from existing tokens (Lipgloss `Lighten`/`Darken`).

## Related Code Files

- Create: `internal/tui/menu/keys.go`
- Modify: `internal/tui/menu/menu.go`, `internal/tui/menu/styles.go`

## Implementation Steps

### Task 3.1 — Number shortcuts

1. In `Update()` `tea.KeyMsg` switch, add cases `"1"`, `"2"`, `"3"`:
   ```go
   case "1", "2", "3":
       idx := int(msg.String()[0] - '1')
       if idx < len(menuItems) {
           a.cursor = idx
           a.selected = Selection(idx + 1)
           return a, tea.Quit
       }
   ```
2. Manually test.

**Commit:** `feat(tui/menu): number shortcuts to jump-select`
**Push:** `git push origin main`

### Task 3.2 — bubbles/help + bubbles/key integration

1. Create `internal/tui/menu/keys.go`:
   ```go
   package menu

   import "charm.land/bubbles/v2/key"

   type keyMap struct {
       Up, Down, Select, Quit, Help key.Binding
       Jump                         key.Binding
   }

   func defaultKeys() keyMap {
       return keyMap{
           Up:     key.NewBinding(key.WithKeys("up", "k"),    key.WithHelp("↑/k", "up")),
           Down:   key.NewBinding(key.WithKeys("down", "j"),  key.WithHelp("↓/j", "down")),
           Select: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("↵", "select")),
           Jump:   key.NewBinding(key.WithKeys("1","2","3"),  key.WithHelp("1-3", "jump")),
           Help:   key.NewBinding(key.WithKeys("?"),          key.WithHelp("?", "help")),
           Quit:   key.NewBinding(key.WithKeys("q","esc"),    key.WithHelp("q", "quit")),
       }
   }

   func (k keyMap) ShortHelp() []key.Binding {
       return []key.Binding{k.Up, k.Down, k.Select, k.Jump, k.Quit, k.Help}
   }
   func (k keyMap) FullHelp() [][]key.Binding {
       return [][]key.Binding{
           {k.Up, k.Down}, {k.Select, k.Jump}, {k.Help, k.Quit},
       }
   }
   ```
2. In `App`, add `keys keyMap` and `help help.Model`. Initialize in `New()`.
3. In `Update()`, route `?` to `a.help.ShowAll = !a.help.ShowAll`. Pass `WindowSizeMsg.Width` to `a.help.Width`.
4. In `View()`, replace `helpStyle.Render(static)` with `a.help.View(a.keys)`.

**Commit:** `feat(tui/menu): integrate bubbles/help with key bindings`
**Push:** `git push origin main`

### Task 3.3 — Adaptive color for text + borders

1. In `styles.go`, swap text colors used in active/inactive titles & descriptions to:
   ```go
   import "charm.land/lipgloss/v2/compat"
   var textBody = compat.AdaptiveColor{
       Light: lipgloss.Color("#1A202C"),
       Dark:  lipgloss.Color(theme.TextStrong),
   }
   ```
2. Apply to `activeTitleStyle`, `inactiveTitleStyle`, descriptions. Keep brand `Primary` solid (purple reads OK on both).
3. Manually verify on dark terminal (default) and a light theme (iTerm `Solarized Light`).

**Commit:** `style(tui/menu): adopt adaptive colors for light/dark terminals`
**Push:** `git push origin main`

## Success Criteria

- [x] Pressing `1`/`2`/`3` selects respective option.
- [x] `?` toggles full help; widths down to 60 cols don't overflow.
- [x] Footer rendered by `bubbles/help`, not hand-built string.
- [x] Text legible on both dark and light terminals.
- [x] 3 commits pushed.
- [x] `go test ./...` green.

## Risk Assessment

- **Risk:** `bubbles/v2/help` API name drift. **Mitigation:** import `charm.land/bubbles/v2/help` and check `pkg.go.dev` doc snapshot before integration.
- **Risk:** `compat.AdaptiveColor` requires terminal background detection that may flicker. **Mitigation:** scope to text only, leave borders solid.

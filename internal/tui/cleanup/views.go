package cleanup

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/pkg"
	"github.com/bavanchun/OmniClean/internal/tui/components"
)

const panelTitle = " Cleanup Suggestions "

// viewLoading renders the spinner shown while aggregating candidates.
func (a *App) viewLoading() string {
	body := fmt.Sprintf("  %s  %s", a.sp.View(), a.theme.Body.Render("Scanning managers for removable packages…"))
	panel := components.Panel(a.theme, panelTitle, body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{{Key: "ctrl+c", Action: "cancel"}})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewList renders the scrollable candidate list with role badges.
func (a *App) viewList() string {
	if len(a.candidates) == 0 {
		body := a.theme.Success.Render("  Nothing to clean up — no orphan or removable leaf packages found.")
		panel := components.Panel(a.theme, panelTitle, body,
			components.PanelOpts{Width: a.safeWidth(), Accent: true})
		footer := components.KeyHints(a.theme, []components.KeyHint{{Key: "q", Action: "quit"}})
		return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
	}

	total := len(a.candidates)
	vr := a.visibleRows()
	end := a.scrollOffset + vr
	if end > total {
		end = total
	}

	var totalSize int64
	selCount := 0
	for _, p := range a.candidates {
		totalSize += p.Size
		if a.selected[key(p)] {
			selCount++
		}
	}

	var lines []string
	for i, p := range a.candidates[a.scrollOffset:end] {
		idx := a.scrollOffset + i

		check := a.theme.Dim.Render("☐")
		if a.selected[key(p)] {
			check = a.theme.Success.Render("☑")
		}
		marker := "  "
		if idx == a.cursor {
			marker = a.theme.Subtitle.Render("➤ ")
		}

		badgeKind := components.BadgeInfo
		if p.Role == pkg.RoleOrphan {
			badgeKind = components.BadgeWarning
		}
		badge := components.Badge(a.theme, badgeKind, " "+roleBadge(p.Role)+" ")

		row := fmt.Sprintf("%s%s  %-28s %s  %-8s  %s",
			marker, check,
			truncate(p.Name, 28),
			badge,
			a.theme.Subtle.Render(string(p.Manager)),
			a.theme.Subtle.Render(installedAtLabel(p)),
		)
		lines = append(lines, row)
	}

	header := fmt.Sprintf("%s— %d candidates  Total: %s  [%d/%d selected] ",
		panelTitle, total, formatBytes(totalSize), selCount, total)
	panel := components.Panel(a.theme, header, strings.Join(lines, "\n"),
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "space", Action: "toggle"},
		{Key: "a", Action: "all"},
		{Key: "enter", Action: "confirm"},
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewConfirm asks the user to confirm removal of the selected candidates.
func (a *App) viewConfirm() string {
	sel := a.selectedPkgs()
	var total int64
	for _, p := range sel {
		total += p.Size
	}

	header := a.theme.Strong.Render(fmt.Sprintf("Remove %d package(s) (%s)?", len(sel), formatBytes(total)))
	kind := components.BadgeWarning
	note := " Packages will be uninstalled via their managers "
	if a.cfg.DryRun {
		kind = components.BadgeDryRun
		note = " DRY RUN — no changes will be made "
	}
	warn := components.Badge(a.theme, kind, note)

	body := lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		warn, "",
		a.theme.Body.Render("  y / enter   confirm and remove"),
		a.theme.Body.Render("  n / esc     back to selection"),
	)
	panel := components.Panel(a.theme, " Confirm cleanup ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "y", Action: "confirm"},
		{Key: "n", Action: "back"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewDeleting shows the spinner while removals run.
func (a *App) viewDeleting() string {
	body := fmt.Sprintf("  %s  %s", a.sp.View(), a.theme.Body.Render("Removing selected packages…"))
	return components.Panel(a.theme, " Cleaning up ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
}

// viewResult renders the post-removal summary.
func (a *App) viewResult() string {
	var freed int64
	var failed int
	var rows []string
	for _, r := range a.results {
		if r.Err != nil {
			failed++
			rows = append(rows, "  "+a.theme.Error.Render("✗")+"  "+
				a.theme.Strong.Render(r.Package.Name)+"  "+a.theme.Error.Render(r.Err.Error()))
			continue
		}
		freed += r.Package.Size
		detail := ""
		if r.DryRunCmd != "" {
			detail = "  " + a.theme.Subtle.Render(r.DryRunCmd)
		} else if r.LeftoverTotal > 0 {
			detail = "  " + a.theme.Subtle.Render(fmt.Sprintf("+%s leftovers", formatBytes(r.LeftoverTotal)))
		}
		rows = append(rows, "  "+a.theme.Success.Render("✓")+"  "+a.theme.Strong.Render(r.Package.Name)+detail)
	}

	chips := []string{
		components.Badge(a.theme, components.BadgeSuccess, fmt.Sprintf(" %d removed ", len(a.results)-failed)),
		components.Badge(a.theme, components.BadgeInfo, fmt.Sprintf(" %s freed ", formatBytes(freed))),
	}
	if failed > 0 {
		chips = append(chips, components.Badge(a.theme, components.BadgeError, fmt.Sprintf(" %d failed ", failed)))
	}
	if a.cfg.DryRun {
		chips = append(chips, components.Badge(a.theme, components.BadgeDryRun, " DRY RUN "))
	}
	chipRow := lipgloss.JoinHorizontal(lipgloss.Left, chips...)

	panel := components.Panel(a.theme, " Cleanup results ", strings.Join(rows, "\n"),
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{{Key: "q", Action: "quit"}})
	return lipgloss.JoinVertical(lipgloss.Left, chipRow, "", panel, "", footer)
}

// installedAtLabel renders the best-effort install time, or "—" when unknown.
func installedAtLabel(p pkg.Package) string {
	if p.InstalledAt.IsZero() {
		return "—"
	}
	return p.InstalledAt.Format("2006-01-02")
}

//go:build darwin

package appuninstall

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/bavanchun/OmniClean/internal/tui/components"
)

// viewLoading renders the spinner shown while scanning for .app bundles.
func (a *App) viewLoading() string {
	body := fmt.Sprintf("  %s  %s",
		a.sp.View(),
		a.theme.Body.Render("Scanning applications…"),
	)
	panel := components.Panel(a.theme, " Uninstall Apps ",
		lipgloss.JoinVertical(lipgloss.Left,
			a.theme.Subtitle.Render("✦  OmniClean uninstall  ✦"), "", body,
		),
		components.PanelOpts{Width: a.safeWidth(), Accent: true},
	)
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "ctrl+c", Action: "cancel"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewList renders the scrollable app-selection list.
func (a *App) viewList() string {
	if len(a.bundles) == 0 {
		body := a.theme.Success.Render("  No applications found in scan roots.")
		panel := components.Panel(a.theme, " Uninstall Apps ", body,
			components.PanelOpts{Width: a.safeWidth(), Accent: true})
		footer := components.KeyHints(a.theme, []components.KeyHint{{Key: "q", Action: "quit"}})
		return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
	}

	total := len(a.bundles)
	a.clampScroll(total)
	vr := a.visibleRows()

	var totalSize int64
	selCount := 0
	for _, b := range a.bundles {
		totalSize += b.Size
		if a.selected[b.Path] {
			selCount++
		}
	}

	end := a.scrollOffset + vr
	if end > total {
		end = total
	}

	var lines []string
	for i, b := range a.bundles[a.scrollOffset:end] {
		idx := a.scrollOffset + i

		check := a.theme.Dim.Render("☐")
		if a.selected[b.Path] {
			check = a.theme.Success.Render("☑")
		}
		marker := "  "
		if idx == a.cursor {
			marker = a.theme.Subtitle.Render("➤ ")
		}

		ver := ""
		if b.Version != "" {
			ver = "  " + a.theme.Subtle.Render(b.Version)
		}
		size := a.theme.Strong.Render(formatBytes(b.Size))
		row := fmt.Sprintf("%s%s  %-32s%s  %s",
			marker, check,
			truncate(b.Name, 32),
			ver,
			size,
		)
		lines = append(lines, row)
	}

	header := fmt.Sprintf(" Uninstall Apps — %d apps  Total: %s  [%d/%d selected] ",
		total, formatBytes(totalSize), selCount, total)
	body := strings.Join(lines, "\n")
	panel := components.Panel(a.theme, header, body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "space", Action: "toggle"},
		{Key: "a", Action: "all"},
		{Key: "enter", Action: "confirm"},
		{Key: "d", Action: "detail"},
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewDetail renders the detail panel for the focused bundle.
func (a *App) viewDetail() string {
	b := a.detailBundle
	check := a.theme.Dim.Render("☐  not selected")
	if a.selected[b.Path] {
		check = a.theme.Success.Render("☑  selected for removal")
	}

	rows := []string{
		fmt.Sprintf("  %-18s %s", a.theme.Subtle.Render("Bundle ID:"), a.theme.Body.Render(orDash(b.BundleID))),
		fmt.Sprintf("  %-18s %s", a.theme.Subtle.Render("Version:"), a.theme.Body.Render(orDash(b.Version))),
		fmt.Sprintf("  %-18s %s", a.theme.Subtle.Render("Size:"), a.theme.Strong.Render(formatBytes(b.Size))),
		fmt.Sprintf("  %-18s %s", a.theme.Subtle.Render("Last Modified:"), a.theme.Body.Render(b.LastModTime.Format("2006-01-02"))),
		fmt.Sprintf("  %-18s %s", a.theme.Subtle.Render("Path:"), a.theme.Subtle.Render(b.Path)),
		"",
		"  " + check,
	}

	body := strings.Join(rows, "\n")
	panel := components.Panel(a.theme, fmt.Sprintf(" %s ", b.Name), body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "space", Action: "toggle selection"},
		{Key: "esc", Action: "back"},
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewConfirmBundle asks the user to confirm deletion of the selected bundles.
func (a *App) viewConfirmBundle() string {
	sel := a.selectedBundles()
	total := totalBundleSize(sel)

	header := a.theme.Strong.Render(
		fmt.Sprintf("Delete %d app(s) (%s)?", len(sel), formatBytes(total)),
	)
	warn := components.Badge(a.theme, components.BadgeWarning, " Apps will be moved to Trash ")

	body := lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		warn, "",
		a.theme.Body.Render("  y / enter   confirm and remove"),
		a.theme.Body.Render("  n / esc     back to selection"),
	)
	panel := components.Panel(a.theme, " Confirm uninstall ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "y", Action: "confirm"},
		{Key: "n", Action: "back"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewDeletingBundle shows the spinner while bundles are being removed.
func (a *App) viewDeletingBundle() string {
	body := fmt.Sprintf("  %s  %s", a.sp.View(), a.theme.Body.Render("Removing selected apps…"))
	panel := components.Panel(a.theme, " Uninstalling ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	return panel
}

// viewLeftoverScan shows the spinner while orphan files are being located.
func (a *App) viewLeftoverScan() string {
	body := fmt.Sprintf("  %s  %s", a.sp.View(), a.theme.Body.Render("Scanning for leftover files…"))
	panel := components.Panel(a.theme, " Leftover scan ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	return panel
}

// viewConfirmLeftovers lists discovered orphan paths and asks whether to delete them.
func (a *App) viewConfirmLeftovers() string {
	total := totalLeftoverSize(a.leftovers)

	header := a.theme.Strong.Render(
		fmt.Sprintf("Found %d leftover path(s) (%s)", len(a.leftovers), formatBytes(total)),
	)
	warn := components.Badge(a.theme, components.BadgeWarning,
		" Permanent deletion — not sent to Trash ")

	var entryLines []string
	for _, e := range a.leftovers {
		desc := ""
		if e.Desc != "" {
			desc = "  " + a.theme.Subtle.Render("["+e.Desc+"]")
		}
		line := fmt.Sprintf("  %s  %s%s",
			a.theme.Body.Render(truncate(e.Path, 60)),
			a.theme.Strong.Render(formatBytes(e.Size)),
			desc,
		)
		entryLines = append(entryLines, line)
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		header, "",
		warn, "",
		strings.Join(entryLines, "\n"),
		"",
		a.theme.Body.Render("  y / enter   delete all leftovers"),
		a.theme.Body.Render("  n / esc     skip leftover cleanup"),
	)
	panel := components.Panel(a.theme, " Leftover files ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "y", Action: "delete leftovers"},
		{Key: "n", Action: "skip"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// viewDeletingLeftovers shows the spinner while orphan files are being removed.
func (a *App) viewDeletingLeftovers() string {
	body := fmt.Sprintf("  %s  %s", a.sp.View(), a.theme.Body.Render("Removing leftover files…"))
	panel := components.Panel(a.theme, " Cleaning up ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	return panel
}

// viewResult renders the post-deletion summary.
func (a *App) viewResult() string {
	// Count bundle outcomes.
	var bundleFreed int64
	var bundleFailed int
	for _, r := range a.bundleResults {
		if r.Err != nil {
			bundleFailed++
		} else {
			for _, b := range a.bundles {
				if b.Path == r.Path {
					bundleFreed += b.Size
					break
				}
			}
		}
	}

	// Count leftover outcomes.
	var leftoverFreed int64
	var leftoverFailed int
	for _, r := range a.leftoverResults {
		if r.Err != nil {
			leftoverFailed++
		} else {
			for _, e := range a.leftovers {
				if e.Path == r.Path {
					leftoverFreed += e.Size
					break
				}
			}
		}
	}

	totalFreed := bundleFreed + leftoverFreed

	// Build chip row.
	chips := []string{
		components.Badge(a.theme, components.BadgeSuccess,
			fmt.Sprintf(" %d app(s) removed ", len(a.deletedBundles))),
		components.Badge(a.theme, components.BadgeInfo,
			fmt.Sprintf(" %s freed ", formatBytes(totalFreed))),
	}
	if bundleFailed+leftoverFailed > 0 {
		chips = append(chips, components.Badge(a.theme, components.BadgeError,
			fmt.Sprintf(" %d failed ", bundleFailed+leftoverFailed)))
	}
	if a.cfg.DryRun {
		chips = append(chips, components.Badge(a.theme, components.BadgeDryRun, " DRY RUN "))
	}
	chipRow := lipgloss.JoinHorizontal(lipgloss.Left, chips...)

	// Bundle result rows.
	var rows []string
	for _, r := range a.bundleResults {
		name := r.Path
		for _, b := range a.bundles {
			if b.Path == r.Path {
				name = b.Name
				break
			}
		}
		if r.Err != nil {
			rows = append(rows, "  "+a.theme.Error.Render("✗")+"  "+
				a.theme.Strong.Render(name)+"  "+a.theme.Error.Render(r.Err.Error()))
		} else {
			size := ""
			for _, b := range a.bundles {
				if b.Path == r.Path {
					size = "  " + a.theme.Subtle.Render(formatBytes(b.Size))
					break
				}
			}
			rows = append(rows, "  "+a.theme.Success.Render("✓")+"  "+
				a.theme.Strong.Render(name)+size)
		}
	}

	// Leftover summary row.
	if len(a.leftoverResults) > 0 {
		cleaned := 0
		for _, r := range a.leftoverResults {
			if r.Err == nil {
				cleaned++
			}
		}
		rows = append(rows, "  "+a.theme.Success.Render("✓")+"  "+
			a.theme.Body.Render(fmt.Sprintf("%d leftover path(s)  %s", cleaned, formatBytes(leftoverFreed))))
	}

	body := strings.Join(rows, "\n")
	panel := components.Panel(a.theme, " Uninstall results ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})

	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, chipRow, "", panel, "", footer)
}

// viewError renders a fatal error (e.g. scan failure).
func (a *App) viewError() string {
	body := a.theme.Error.Render("Error: " + a.scanErr.Error())
	panel := components.Panel(a.theme, " Error ", body,
		components.PanelOpts{Width: a.safeWidth(), Accent: true})
	footer := components.KeyHints(a.theme, []components.KeyHint{
		{Key: "q", Action: "quit"},
	})
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", footer)
}

// orDash returns s if non-empty, otherwise "—".
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

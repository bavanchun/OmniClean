package tui

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// listModel wraps bubbles/list and adds multi-select state.
type listModel struct {
	list          list.Model
	selected      map[string]bool // key: "manager:name"
	styles        Styles
	warnings      []string
	originalItems []pkg.Package // preserved for sort reset
	sortedBySize  bool
}

// packageDelegate renders each package row with manager badge and selection checkbox.
type packageDelegate struct {
	styles   Styles
	selected map[string]bool // shared reference — Go maps are reference types
}

func (d packageDelegate) Height() int                               { return 1 }
func (d packageDelegate) Spacing() int                              { return 0 }
func (d packageDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d packageDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	p, ok := item.(pkg.Package)
	if !ok {
		return
	}

	isCursor := index == m.Index()
	isChecked := d.selected[selectionKey(p)]

	cursor := "  "
	if isCursor {
		cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Render("> ")
	}

	checkbox := "[ ]"
	if isChecked {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("#48BB78")).Bold(true).Render("[✓]")
	}

	name := p.Name
	if isCursor {
		name = d.styles.Selected.Render(name)
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#718096"))
	badge := d.styles.BadgeFor(string(p.Manager))
	version := dimStyle.Render(p.Version)

	if p.Size > 0 {
		sizeStr := dimStyle.Render(formatBytes(p.Size))
		fmt.Fprintf(w, "%s%s %s %s %s  %s", cursor, checkbox, name, badge, version, sizeStr)
	} else {
		fmt.Fprintf(w, "%s%s %s %s %s", cursor, checkbox, name, badge, version)
	}
}

func newListModel(packages []pkg.Package, styles Styles, warnings []string) listModel {
	items := packagesToItems(packages)

	selected := make(map[string]bool)
	delegate := packageDelegate{styles: styles, selected: selected}

	l := list.New(items, delegate, 0, 0)
	l.Title = "OmniClean"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title

	return listModel{
		list:          l,
		selected:      selected,
		styles:        styles,
		warnings:      warnings,
		originalItems: packages,
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			if item, ok := m.list.SelectedItem().(pkg.Package); ok {
				key := selectionKey(item)
				m.selected[key] = !m.selected[key]
			}
			return m, nil
		case "s":
			m.sortedBySize = !m.sortedBySize
			if m.sortedBySize {
				sorted := make([]pkg.Package, len(m.originalItems))
				copy(sorted, m.originalItems)
				sort.SliceStable(sorted, func(i, j int) bool {
					return sorted[i].Size > sorted[j].Size
				})
				m.list.SetItems(packagesToItems(sorted))
			} else {
				m.list.SetItems(packagesToItems(m.originalItems))
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		warningLines := len(m.warnings)
		m.list.SetSize(msg.Width, msg.Height-4-warningLines)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	selectedCount := 0
	for _, v := range m.selected {
		if v {
			selectedCount++
		}
	}

	var parts []string
	parts = append(parts, m.list.View())

	// Show warning bar if any detectors failed
	if len(m.warnings) > 0 {
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F6AD55"))
		parts = append(parts, warnStyle.Render("  ⚠ Some detectors failed: "+strings.Join(m.warnings, ", ")))
	}

	sortHint := "s: sort size"
	if m.sortedBySize {
		sortHint = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Render("s: sort size ↓")
	}
	status := m.styles.StatusBar.Render(
		fmt.Sprintf("%d selected  ·  space: select  %s  d: details  enter: confirm  q: quit", selectedCount, sortHint),
	)
	parts = append(parts, status)

	return strings.Join(parts, "\n")
}

// SelectedPackages returns all packages the user has toggled.
func (m listModel) SelectedPackages() []pkg.Package {
	allItems := m.list.Items()
	var result []pkg.Package
	for _, item := range allItems {
		p, ok := item.(pkg.Package)
		if !ok {
			continue
		}
		if m.selected[selectionKey(p)] {
			result = append(result, p)
		}
	}
	return result
}

func selectionKey(p pkg.Package) string {
	return string(p.Manager) + ":" + p.Name
}

func packagesToItems(packages []pkg.Package) []list.Item {
	items := make([]list.Item, len(packages))
	for i, p := range packages {
		items[i] = p
	}
	return items
}

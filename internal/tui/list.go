package tui

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/truncate"

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
	keys          KeyMap
	help          help.Model
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

	// Checkbox & Cursor
	cursor := "  "
	if isCursor {
		cursor = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Render("▸ ")
	}

	checkbox := "[ ]"
	if isChecked {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Bold(true).Render("[✓]")
	}
	stateCol := fmt.Sprintf("%s%s ", cursor, checkbox) // approx 6 chars width

	// Badge (ColWidthBadge)
	badgeCol := lipgloss.NewStyle().Width(ColWidthBadge).Render(d.styles.BadgeFor(string(p.Manager)))

	// Size (ColWidthSize) - Right aligned
	sizeStr := ""
	if p.Size > 0 {
		sizeStr = formatBytes(p.Size)
	}
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim))
	sizeCol := lipgloss.NewStyle().Width(ColWidthSize).Align(lipgloss.Right).Render(dimStyle.Render(sizeStr))

	// Version (ColWidthVersion) - Truncate if too long
	versionText := p.Version
	if len(versionText) > ColWidthVersion {
		versionText = truncate.StringWithTail(versionText, uint(ColWidthVersion), "…")
	}
	versionCol := lipgloss.NewStyle().Width(ColWidthVersion).Render(dimStyle.Render(versionText))

	// Name (Flexible Width)
	const listChromeWidth = 4 // Margins/padding added by bubbles/list
	fixedWidths := 6 /* state */ + ColWidthBadge + ColWidthVersion + ColWidthSize + 4 /* spaces */
	
	nameWidth := m.Width() - fixedWidths - listChromeWidth
	if nameWidth < 10 {
		nameWidth = 10
	}

	nameText := p.Name
	if len(nameText) > nameWidth {
		nameText = truncate.StringWithTail(nameText, uint(nameWidth), "…")
	}

	nameStyle := lipgloss.NewStyle()
	if isCursor {
		nameStyle = d.styles.SelectedText
	}
	nameCol := lipgloss.NewStyle().Width(nameWidth).Render(nameStyle.Render(nameText))

	// Assemble Row
	row := lipgloss.JoinHorizontal(lipgloss.Left,
		stateCol,
		badgeCol,
		nameCol,
		"  ",
		versionCol,
		"  ",
		sizeCol,
	)

	// In Bubbles/List, applying background to full width requires adding a padding fill
	if isCursor {
		fill := max(0, m.Width()-listChromeWidth-lipgloss.Width(row))
		row = row + strings.Repeat(" ", fill)
		row = d.styles.SelectedRow.Render(row)
	}

	fmt.Fprint(w, row)
}

func newListModel(packages []pkg.Package, styles Styles, warnings []string) listModel {
	items := packagesToItems(packages)

	selected := make(map[string]bool)
	delegate := packageDelegate{
		styles: styles, selected: selected,
	}

	l := list.New(items, delegate, 0, 0)
	l.Title = "✦ OmniClean"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title

	keys := DefaultKeyMap()
	h := help.New()
	h.Styles = help.DefaultDarkStyles()

	return listModel{
		list:          l,
		selected:      selected,
		styles:        styles,
		warnings:      warnings,
		originalItems: packages,
		keys:          keys,
		help:          h,
	}
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Select):
			if item, ok := m.list.SelectedItem().(pkg.Package); ok {
				k := selectionKey(item)
				m.selected[k] = !m.selected[k]
			}
			return m, nil
		case key.Matches(msg, m.keys.SelectAll):
			allSelected := true
			for _, item := range m.list.Items() {
				p, ok := item.(pkg.Package)
				if !ok {
					continue
				}
				if !m.selected[selectionKey(p)] {
					allSelected = false
					break
				}
			}
			for _, item := range m.list.Items() {
				p, ok := item.(pkg.Package)
				if !ok {
					continue
				}
				m.selected[selectionKey(p)] = !allSelected
			}
			return m, nil
		case key.Matches(msg, m.keys.SortSize):
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
		case msg.String() == "?":
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}
	case tea.WindowSizeMsg:
		warningLines := len(m.warnings)
		m.list.SetSize(msg.Width, msg.Height-6-warningLines)
		m.help.SetWidth(msg.Width)
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
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning))
		parts = append(parts, warnStyle.Render("  ⚠ Some detectors failed: "+strings.Join(m.warnings, ", ")))
	}

	// Selection status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(ColorPrimary)).
		Padding(0, 1).
		Bold(true)
		
	sortIndicator := ""
	if m.sortedBySize {
		sortIndicator = " ↓size"
	}
	
	statusText := fmt.Sprintf("%d selected%s", selectedCount, sortIndicator)
	leftStatus := statusStyle.Render(statusText)
	
	// Create a full-width status line using the selected row background color
	listWidth := m.list.Width()
	if listWidth == 0 {
		listWidth = 80 // fallback
	}
	
	fillWidth := max(0, listWidth - lipgloss.Width(leftStatus))
	fill := lipgloss.NewStyle().Background(lipgloss.Color(ColorSelectedBg)).Render(strings.Repeat(" ", fillWidth))
	
	statusBar := lipgloss.JoinHorizontal(lipgloss.Left, leftStatus, fill)

	// Help component
	helpView := m.help.View(m.keys)

	parts = append(parts, statusBar)
	parts = append(parts, "  "+helpView)

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

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
	styles       Styles
	selected     map[string]bool // shared reference — Go maps are reference types
	nameWidth    int             // max name length across all packages
	versionWidth int             // max version length
	badgeWidth   int             // max badge length e.g. len("[flatpak]")=9
}

// calcColumnWidths returns the max name, version, and badge widths across packages.
func calcColumnWidths(packages []pkg.Package) (nameW, versionW, badgeW int) {
	for _, p := range packages {
		if n := len(p.Name); n > nameW {
			nameW = n
		}
		if v := len(p.Version); v > versionW {
			versionW = v
		}
		if b := len(string(p.Manager)) + 2; b > badgeW { // "[manager]"
			badgeW = b
		}
	}
	return
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
		cursor = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Render("▸ ")
	}

	checkbox := "[ ]"
	if isChecked {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Bold(true).Render("[✓]")
	}

	// Name column — style first, then pad the remaining space with plain spaces.
	namePad := strings.Repeat(" ", max(0, d.nameWidth-len(p.Name)))
	var nameCol string
	if isCursor {
		nameCol = d.styles.Selected.Render(p.Name) + namePad
	} else {
		nameCol = p.Name + namePad
	}

	// Badge column — pad raw width, then apply color.
	badgePad := strings.Repeat(" ", max(0, d.badgeWidth-(len(string(p.Manager))+2)))
	badge := d.styles.BadgeFor(string(p.Manager)) + badgePad

	// Version column — pad then dim.
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDim))
	verPad := p.Version + strings.Repeat(" ", max(0, d.versionWidth-len(p.Version)))
	version := dimStyle.Render(verPad)

	// Size column — right-aligned to 10 chars.
	if p.Size > 0 {
		sizeStr := formatBytes(p.Size)
		sizePad := strings.Repeat(" ", max(0, 10-len(sizeStr)))
		fmt.Fprintf(w, "%s%s %s  %s  %s  %s%s", cursor, checkbox, nameCol, badge, version, sizePad, dimStyle.Render(sizeStr))
	} else {
		fmt.Fprintf(w, "%s%s %s  %s  %s", cursor, checkbox, nameCol, badge, version)
	}
}

func newListModel(packages []pkg.Package, styles Styles, warnings []string) listModel {
	items := packagesToItems(packages)

	selected := make(map[string]bool)
	nameW, versionW, badgeW := calcColumnWidths(packages)
	delegate := packageDelegate{
		styles: styles, selected: selected,
		nameWidth: nameW, versionWidth: versionW, badgeWidth: badgeW,
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

	// Selection counter
	counterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Bold(true)
	sortIndicator := ""
	if m.sortedBySize {
		sortIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimary)).Render(" ↓size")
	}
	counter := counterStyle.Render(fmt.Sprintf("  %d selected", selectedCount)) + sortIndicator

	// Help component
	helpView := m.help.View(m.keys)

	parts = append(parts, counter)
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

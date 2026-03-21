package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bavanchun/OmniClean/internal/pkg"
)

// listModel wraps bubbles/list and adds multi-select state.
type listModel struct {
	list     list.Model
	selected map[string]bool // key: package name
	styles   Styles
}

// packageDelegate renders each package row with a manager badge.
type packageDelegate struct {
	styles Styles
}

func (d packageDelegate) Height() int                               { return 1 }
func (d packageDelegate) Spacing() int                             { return 0 }
func (d packageDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d packageDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	p, ok := item.(pkg.Package)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	badge := d.styles.BadgeFor(string(p.Manager))

	name := p.Name
	if isSelected {
		name = d.styles.Selected.Render("> " + name)
	} else {
		name = "  " + name
	}

	version := lipgloss.NewStyle().Foreground(lipgloss.Color("#718096")).Render(p.Version)
	line := fmt.Sprintf("%s %s %s", name, badge, version)
	fmt.Fprintln(w, line)
}

func newListModel(packages []pkg.Package, styles Styles) listModel {
	items := make([]list.Item, len(packages))
	for i, p := range packages {
		items[i] = p
	}

	delegate := packageDelegate{styles: styles}
	l := list.New(items, delegate, 0, 0)
	l.Title = "OmniClean"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title

	return listModel{
		list:     l,
		selected: make(map[string]bool),
		styles:   styles,
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
			// Toggle selection on the current item
			if item, ok := m.list.SelectedItem().(pkg.Package); ok {
				key := selectionKey(item)
				m.selected[key] = !m.selected[key]
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-4) // leave room for header + help
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

	status := m.styles.StatusBar.Render(
		fmt.Sprintf("%d selected | space: select  enter: confirm  q: quit", selectedCount),
	)

	return m.list.View() + "\n" + status
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
	return strings.Join([]string{string(p.Manager), p.Name}, ":")
}

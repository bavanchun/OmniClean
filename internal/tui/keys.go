package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// KeyMap defines all key bindings for the application.
// Used by both key handling and the help component.
type KeyMap struct {
	Quit      key.Binding
	ForceQuit key.Binding
	Select    key.Binding
	SelectAll key.Binding
	Confirm   key.Binding
	Detail    key.Binding
	SortSize  key.Binding
	Back      key.Binding
	Refresh   key.Binding
	Settings  key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "force quit"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Detail: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "details"),
		),
		SortSize: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort by size"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Settings: key.NewBinding(
			key.WithKeys(","),
			key.WithHelp(",", "settings"),
		),
	}
}

// ShortHelp returns the short help key bindings for the list view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.SortSize, k.Detail, k.Confirm, k.Settings, k.Quit}
}

// FullHelp returns the full help key bindings organized in columns.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.SelectAll, k.SortSize},
		{k.Detail, k.Confirm, k.Back},
		{k.Settings, k.Refresh, k.Quit, k.ForceQuit},
	}
}

// Ensure KeyMap implements help.KeyMap.
var _ help.KeyMap = KeyMap{}

// ResultKeyMap defines key bindings for the result view.
type ResultKeyMap struct {
	Refresh key.Binding
	Quit    key.Binding
}

// ShortHelp returns help for result view.
func (k ResultKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Refresh, k.Quit}
}

// FullHelp returns full help for result view.
func (k ResultKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Refresh, k.Quit}}
}

// DefaultResultKeyMap returns key bindings for the result view.
func DefaultResultKeyMap() ResultKeyMap {
	return ResultKeyMap{
		Refresh: key.NewBinding(
			key.WithKeys("enter", "r"),
			key.WithHelp("enter/r", "back to list"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}

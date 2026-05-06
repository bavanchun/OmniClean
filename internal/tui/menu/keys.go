package menu

import "charm.land/bubbles/v2/key"

// keyMap binds the menu's input keys to their help-bar labels.
// It implements help.KeyMap (ShortHelp + FullHelp) so bubbles/help
// can render and gracefully truncate the footer at narrow widths.
type keyMap struct {
	Up, Down key.Binding
	Select   key.Binding
	Jump     key.Binding
	Help     key.Binding
	Quit     key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Select: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("↵", "select")),
		Jump:   key.NewBinding(key.WithKeys("1", "2", "3"), key.WithHelp("1-3", "jump")),
		Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:   key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Jump, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Select, k.Jump},
		{k.Help, k.Quit},
	}
}

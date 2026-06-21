package cleanup

// keyMap holds the cleanup TUI key bindings. Mirrors the appuninstall keymap
// precedent (named constants instead of scattered string literals).
type keyMap struct {
	Up    []string
	Down  []string
	Space string
	All   string
	Enter string
	Back  string
	Yes   string
	No    string
	Quit  []string
}

func defaultKeys() keyMap {
	return keyMap{
		Up:    []string{"up", "k"},
		Down:  []string{"down", "j"},
		Space: "space",
		All:   "a",
		Enter: "enter",
		Back:  "esc",
		Yes:   "y",
		No:    "n",
		Quit:  []string{"q", "ctrl+c"},
	}
}

func inList(key string, list []string) bool {
	for _, k := range list {
		if k == key {
			return true
		}
	}
	return false
}

func (km keyMap) isUp(key string) bool   { return inList(key, km.Up) }
func (km keyMap) isDown(key string) bool { return inList(key, km.Down) }
func (km keyMap) isQuit(key string) bool { return inList(key, km.Quit) }

//go:build darwin

package appuninstall

// keyMap holds the string representations of all key bindings used across
// the appuninstall TUI. Using named constants avoids scatter of raw string
// literals throughout the Update / view code.
type keyMap struct {
	Up    []string
	Down  []string
	Space string
	All   string
	Enter string
	Detail string
	Back  string
	Yes   string
	No    string
	Quit  []string
}

// defaultKeys returns the canonical key map for the appuninstall TUI.
func defaultKeys() keyMap {
	return keyMap{
		Up:     []string{"up", "k"},
		Down:   []string{"down", "j"},
		Space:  " ",
		All:    "a",
		Enter:  "enter",
		Detail: "d",
		Back:   "esc",
		Yes:    "y",
		No:     "n",
		Quit:   []string{"q", "ctrl+c"},
	}
}

// isUp reports whether key matches the Up binding.
func (km keyMap) isUp(key string) bool {
	for _, k := range km.Up {
		if k == key {
			return true
		}
	}
	return false
}

// isDown reports whether key matches the Down binding.
func (km keyMap) isDown(key string) bool {
	for _, k := range km.Down {
		if k == key {
			return true
		}
	}
	return false
}

// isQuit reports whether key matches any Quit binding.
func (km keyMap) isQuit(key string) bool {
	for _, k := range km.Quit {
		if k == key {
			return true
		}
	}
	return false
}

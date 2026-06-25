package detector

// AllDetectors returns every detector OmniClean knows about.
// The list is explicit (not init-based) to keep things testable and readable.
// Detectors for unavailable managers are safe to include — Available() filters them at runtime.
func AllDetectors() []Detector {
	return []Detector{
		// Linux
		NewAPT(DefaultRunner),
		NewDNF(DefaultRunner),
		NewPacman(DefaultRunner),
		NewZypper(DefaultRunner),
		NewSnap(DefaultRunner),
		NewFlatpak(DefaultRunner),
		// macOS + Linux
		NewBrew(DefaultRunner),
		// Cross-platform (language package managers)
		NewPip(DefaultRunner),
		NewNPM(DefaultRunner),
		NewCargo(DefaultRunner),
		// Windows
		NewWinget(DefaultRunner),
		NewChoco(DefaultRunner),
		NewScoop(DefaultRunner),
	}
}

// AvailableDetectors returns only the detectors whose package manager
// is installed on the current system.
func AvailableDetectors() []Detector {
	all := AllDetectors()
	available := make([]Detector, 0, len(all))
	for _, d := range all {
		if d.Available() {
			available = append(available, d)
		}
	}
	return available
}

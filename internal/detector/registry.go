package detector

// AllDetectors returns every detector OmniClean knows about.
// The list is explicit (not init-based) to keep things testable and readable.
func AllDetectors() []Detector {
	return []Detector{
		NewAPT(DefaultRunner),
		NewBrew(DefaultRunner),
		NewSnap(DefaultRunner),
		NewFlatpak(DefaultRunner),
		NewPip(DefaultRunner),
		NewNPM(DefaultRunner),
		NewCargo(DefaultRunner),
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

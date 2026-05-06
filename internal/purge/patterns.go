package purge

// Pattern is a directory name to match plus the stack it belongs to.
// We deliberately match by exact basename rather than glob: this keeps
// scanning fast and makes false positives easy to audit.
type Pattern struct {
	Name  string
	Stack Stack
}

// DefaultPatterns is the canonical list of directory names that are
// safe to purge across stacks. Entries marked as StackGeneric are not
// tied to a specific ecosystem (e.g. "build", "dist").
var DefaultPatterns = []Pattern{
	{Name: "node_modules", Stack: StackNode},
	{Name: ".next", Stack: StackNode},
	{Name: ".nuxt", Stack: StackNode},
	{Name: ".turbo", Stack: StackNode},
	{Name: ".parcel-cache", Stack: StackNode},

	{Name: "target", Stack: StackRust},

	{Name: ".venv", Stack: StackPython},
	{Name: "venv", Stack: StackPython},
	{Name: "__pycache__", Stack: StackPython},
	{Name: ".tox", Stack: StackPython},
	{Name: ".pytest_cache", Stack: StackPython},
	{Name: ".mypy_cache", Stack: StackPython},
	{Name: ".ruff_cache", Stack: StackPython},

	{Name: "vendor", Stack: StackGo},

	{Name: ".gradle", Stack: StackJava},
	{Name: "build", Stack: StackJava},

	{Name: "bin", Stack: StackDotNet},
	{Name: "obj", Stack: StackDotNet},

	{Name: "dist", Stack: StackGeneric},
	{Name: ".cache", Stack: StackGeneric},
}

// matchPattern returns the matching Pattern for a basename, or nil.
func matchPattern(name string) *Pattern {
	for i := range DefaultPatterns {
		if DefaultPatterns[i].Name == name {
			return &DefaultPatterns[i]
		}
	}
	return nil
}

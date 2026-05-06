// Package purge finds and removes ephemeral build artifacts produced by
// project workflows (node_modules, target/, .venv/, ...). It is a peer
// of the leftover package: leftover deals with package manager residue,
// purge deals with project-level disposable directories.
package purge

import "time"

// Stack identifies which language ecosystem a target belongs to. Used
// to render filterable badges in the TUI and to allow per-stack pattern
// toggles in the future.
type Stack string

const (
	StackNode    Stack = "node"
	StackRust    Stack = "rust"
	StackPython  Stack = "python"
	StackGo      Stack = "go"
	StackJava    Stack = "java"
	StackDotNet  Stack = "dotnet"
	StackGeneric Stack = "build"
)

// Target is a single artifact directory or file the scanner identified
// as safe to delete. Project is the human-readable parent project name
// (basename of the project root) used for grouping in the TUI.
type Target struct {
	Path     string
	Project  string
	Stack    Stack
	Pattern  string
	Size     int64
	Modified time.Time
	// Recent is true when Modified is within the recency threshold
	// (default 7 days). Recent targets are pre-unselected in the TUI so
	// users do not nuke an in-progress build by accident.
	Recent bool
}

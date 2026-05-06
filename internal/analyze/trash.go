package analyze

// MoveToTrash sends path to the OS recycle bin / trash. The actual
// implementation is selected at compile time via build tags in
// trash_darwin.go, trash_linux.go, and trash_windows.go.
//
// We deliberately do NOT fall back to os.RemoveAll: skipping the trash
// would silently change the destructive behaviour the user expects, so
// implementations return an error when no trash mechanism is available.

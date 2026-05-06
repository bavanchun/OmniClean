package analyze

// HistoryEntry caches a previously analyzed Result along with the
// cursor position so back/forward navigation feels instant.
type HistoryEntry struct {
	Result      Result
	Cursor      int
	Offset      int
	LargeCursor int
	LargeOffset int
}

// History is a simple bounded stack of HistoryEntry values used by the
// TUI to navigate up/down through the directory tree.
type History struct {
	stack []HistoryEntry
	max   int
}

// NewHistory returns a History bounded to maxEntries (default 32 when 0).
func NewHistory(maxEntries int) *History {
	if maxEntries <= 0 {
		maxEntries = 32
	}
	return &History{max: maxEntries}
}

func (h *History) Push(e HistoryEntry) {
	h.stack = append(h.stack, e)
	if len(h.stack) > h.max {
		h.stack = h.stack[len(h.stack)-h.max:]
	}
}

func (h *History) Pop() (HistoryEntry, bool) {
	if len(h.stack) == 0 {
		return HistoryEntry{}, false
	}
	last := h.stack[len(h.stack)-1]
	h.stack = h.stack[:len(h.stack)-1]
	return last, true
}

func (h *History) Len() int { return len(h.stack) }

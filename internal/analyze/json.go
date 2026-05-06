package analyze

import (
	"encoding/json"
	"io"
)

// WriteJSON serializes res to w using two-space indentation. The struct
// tags on Result/DirEntry/FileEntry shape the payload so consumers see
// stable snake_case keys regardless of internal field renames.
func WriteJSON(w io.Writer, res Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(res)
}

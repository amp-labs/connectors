package utils

import (
	"encoding/json"
	"io"
)

// DumpJSON dumps the given value as JSON to the given writer.
func DumpJSON(v any, w io.Writer) {
	bts, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		Fail("error marshaling to JSON: %w", "error", err)
	}

	_, err = w.Write(append(bts, []byte("\n")...))
	if err != nil {
		Fail("error writing to writer: %w", "error", err)
	}
}

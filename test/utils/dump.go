package utils

import (
	"encoding/json"
	"io"
)

func DumpJSON(v any, w io.Writer) {
	bts, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		Fail("error marshaling to JSON: %w", "error", err)
	}

	_, err = w.Write(bts)
	if err != nil {
		Fail("error writing to writer: %w", "error", err)
	}
}

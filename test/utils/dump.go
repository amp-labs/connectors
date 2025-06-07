package utils

import (
	"encoding/json"
	"fmt"
	"io"
)

// DumpJSON dumps the given value as JSON to the given writer.
func DumpJSON(v any, w io.Writer) {
	encoder := json.NewEncoder(w)

	// JSON may have URLs with special symbols which shouldn't be escaped. Ex: `&`.
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(v); err != nil {
		Fail("error marshaling to JSON: %w", "error", err)
	}
}

func DumpErrorsMap(registry map[string]error, w io.Writer) {
	if len(registry) != 0 {
		_, _ = w.Write([]byte("Errors map is not empty:\n"))
	}

	for key, value := range registry {
		_, _ = w.Write([]byte(fmt.Sprintf("[%v] => %v\n", key, value)))
	}
}

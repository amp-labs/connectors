package utils

import (
	"encoding/json"
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

func PrettyFormatStruct(s any) (string, error) {
	json, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}

	return string(json), nil
}

package utils

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/amp-labs/connectors/common"
)

// DumpJSON dumps the given value as JSON to the given writer.
func DumpJSON(v any, w io.Writer) {
	if result, ok := v.(*common.ListObjectMetadataResult); ok {
		// Nested errors must be explicitly converted to a string to be displayed.
		errorsMap := map[string]string{}

		for k, err := range result.Errors {
			if err != nil {
				errorsMap[k] = err.Error() // convert error to string
			} else {
				errorsMap[k] = ""
			}
		}

		v = map[string]any{
			"Result": result.Result,
			"Errors": errorsMap,
		}
	}

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

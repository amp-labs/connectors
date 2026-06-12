package httpkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ReadJSONBody parses the request body into type T and restores req.Body
// so it can be read again by other handlers.
func ReadJSONBody[T any](req *http.Request) (T, error) {
	var result T

	// Read entire body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read body: %w", err)
	}

	// Restore req.Body with a re-readable copy
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse JSON
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return result, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return result, nil
}

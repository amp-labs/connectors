package codec

import "encoding/json"

// Parse converts any input to type T using type assertion or JSON marshaling.
// Supports primitives, structs, and JSON-compatible types. Returns zero value of T on error.
func Parse[T any](input any) (T, error) {
	if object, ok := input.(T); ok {
		return object, nil
	}

	bytes, err := json.Marshal(input)
	if err != nil {
		var empty T

		return empty, err
	}

	var object T
	if err = json.Unmarshal(bytes, &object); err != nil {
		return object, err
	}

	return object, nil
}

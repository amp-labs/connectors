package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// DumpJSON dumps the given value as JSON to the given writer.
func DumpJSON(v any, w io.Writer) {
	// Convert any error interfaces recursively before encoding.
	convertedValue := substituteErrorsToStrings(v)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(convertedValue); err != nil {
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

// substituteErrorsToStrings recursively converts any Go value into a JSON-safe
// representation (maps, slices, primitives), replacing all errors with strings.
//
// Structs become map[string]any, slices/arrays become []any, maps are preserved
// with converted keys, and all nested values are processed recursively.
func substituteErrorsToStrings(v any) any {
	if v == nil {
		return nil
	}

	// Convert errors to their string form.
	if err, ok := v.(error); ok {
		return err.Error()
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return nil
		}

		return substituteErrorsToStrings(rv.Elem().Interface())
	case reflect.Struct:
		out := make(map[string]any)
		rt := rv.Type()

		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.PkgPath != "" { // unexported
				continue
			}

			out[field.Name] = substituteErrorsToStrings(rv.Field(i).Interface())
		}

		return out
	case reflect.Map:
		out := make(map[string]any)

		for _, key := range rv.MapKeys() {
			// Only string keys are valid JSON object keys; fallback to fmt.Sprint
			k := fmt.Sprint(key.Interface())
			out[k] = substituteErrorsToStrings(rv.MapIndex(key).Interface())
		}

		return out
	case reflect.Slice, reflect.Array:
		n := rv.Len()
		out := make([]any, n)

		for i := 0; i < n; i++ {
			out[i] = substituteErrorsToStrings(rv.Index(i).Interface())
		}

		return out
	default:
		return v
	}
}

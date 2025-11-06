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

// substituteErrorsToStrings recursively traverses an arbitrary Go value and
// replaces every occurrence of an error with its string representation via Error().
//
// The function returns a deep copy of the input value; the original data
// remains unmodified. It is safe to call on any structure intended for logging, or debugging.
func substituteErrorsToStrings(v any) any {
	if v == nil {
		return nil
	}

	// If the top-level value is an error, just return its string
	// That's the goal of this recursive function.
	if err, ok := v.(error); ok {
		return err.Error()
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return v
		}

		// Unwrap and recurse
		return substituteErrorsToStrings(rv.Elem().Interface())
	case reflect.Struct:
		newVal := reflect.New(rv.Type()).Elem()
		for i := 0; i < rv.NumField(); i++ {
			fieldVal := rv.Field(i)
			fieldType := rv.Type().Field(i)

			// Skip unexported fields to avoid panic
			if fieldType.PkgPath != "" {
				continue
			}

			converted := substituteErrorsToStrings(fieldVal.Interface())
			newVal.Field(i).Set(reflect.ValueOf(converted))
		}

		return newVal.Interface()
	case reflect.Slice, reflect.Array:
		n := rv.Len()
		newSlice := reflect.MakeSlice(rv.Type(), n, n)
		for i := 0; i < n; i++ {
			converted := substituteErrorsToStrings(rv.Index(i).Interface())
			newSlice.Index(i).Set(reflect.ValueOf(converted))
		}

		return newSlice.Interface()
	case reflect.Map:
		newMap := reflect.MakeMap(rv.Type())
		for _, key := range rv.MapKeys() {
			converted := substituteErrorsToStrings(rv.MapIndex(key).Interface())
			newMap.SetMapIndex(key, reflect.ValueOf(converted))
		}

		return newMap.Interface()
	default:
		return v
	}
}

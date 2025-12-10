package substitutions

import (
	"reflect"
	"strings"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template
)

// substituteStruct applies substitutions to all string fields in the struct pointed to by input.
// It handles nested structs, pointers, slices, arrays, maps (including pointers-to-maps), and structs inside maps.
func substituteStruct(input any, substitutions map[string]string) error {
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return nil
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		// skip unexported or unsettable fields
		if !field.CanSet() {
			continue
		}

		structField := typ.Field(i)

		skip, ok := structField.Tag.Lookup("skipSubstitutions")
		if ok && skip == "true" {
			continue
		}

		if err := walkValue(field, substitutions); err != nil {
			return err
		}
	}

	return nil
}

// recursing into structs, pointers, slices, arrays, and maps (including values that are structs or pointers).
func walkValue(val reflect.Value, substitutions map[string]string) error { //nolint:cyclop
	switch val.Kind() { //nolint:exhaustive
	case reflect.String:
		s, err := substitute(val.String(), substitutions)
		if err != nil {
			return err
		}

		val.SetString(s)

	case reflect.Pointer:
		if val.IsNil() {
			return nil
		}
		// unwrap pointers uniformly
		return walkValue(val.Elem(), substitutions)

	case reflect.Struct:
		// recurse into nested struct
		return substituteStruct(val.Addr().Interface(), substitutions)

	case reflect.Slice, reflect.Array:
		for i := range val.Len() {
			if err := walkValue(val.Index(i), substitutions); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range val.MapKeys() {
			orig := val.MapIndex(key)
			// wrap non-pointer values in a pointer so walkValue can mutate them
			var ptr reflect.Value
			if orig.Kind() == reflect.Pointer {
				ptr = orig
			} else {
				ptr = reflect.New(orig.Type())
				ptr.Elem().Set(orig)
			}
			// recurse into the wrapped value (handles strings, structs, nested maps, slices, etc.)
			if err := walkValue(ptr, substitutions); err != nil {
				return err
			}
			// write back for non-pointer entries (pointer entries are updated in place)
			if orig.Kind() != reflect.Pointer {
				val.SetMapIndex(key, ptr.Elem())
			}
		}
	default:
		return nil
	}

	return nil
}

// substitute applies text/template substitution to the input string.
func substitute(input string, substitutions map[string]string) (string, error) {
	tmpl, err := template.New("-").Option("missingkey=error").Parse(input)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, substitutions); err != nil {
		return "", err
	}

	return sb.String(), nil
}

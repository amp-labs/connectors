package substitutions

import (
	"reflect"
	"strings"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template
)

// substituteStruct applies substitutions to all string fields in the struct pointed to by input.
// It handles nested structs, pointers, slices, arrays, maps (including pointers-to-maps), and structs inside maps.
func substituteStruct(input interface{}, substitutions map[string]string) error {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// skip unexported or unsettable fields
		if !field.CanSet() {
			continue
		}
		if err := walkValue(field, substitutions); err != nil {
			return err
		}
	}
	return nil
}

// walkValue recursively walks v, handling substitutions for strings,
// recursing into structs, pointers, slices, arrays, and maps (including values that are structs or pointers)
func walkValue(v reflect.Value, substitutions map[string]string) error {
	switch v.Kind() {
	case reflect.String:
		s, err := substitute(v.String(), substitutions)
		if err != nil {
			return err
		}
		v.SetString(s)

	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		// unwrap pointers uniformly
		return walkValue(v.Elem(), substitutions)

	case reflect.Struct:
		// recurse into nested struct
		return substituteStruct(v.Addr().Interface(), substitutions)

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := walkValue(v.Index(i), substitutions); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range v.MapKeys() {
			orig := v.MapIndex(key)
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
				v.SetMapIndex(key, ptr.Elem())
			}
		}
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

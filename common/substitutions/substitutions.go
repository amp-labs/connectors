package substitutions

import (
	"reflect"
	"strings"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template
)

// substituteStruct performs string substitution on the fields of the input struct
// using the substitutions map.
func substituteStruct(input interface{}, substitutions map[string]string) (err error) { //nolint:gocognit,cyclop,lll
	configStruct := reflect.ValueOf(input).Elem()

	for i := range configStruct.NumField() {
		field := configStruct.Field(i)

		// If the field is a string, perform substitution on it.
		if field.Kind() == reflect.String {
			substitutedVal, err := substitute(field.String(), substitutions)
			if err != nil {
				return err
			}

			field.SetString(substitutedVal)
		}

		if field.Kind() == reflect.Pointer {
			if field.Elem().Kind() == reflect.Struct {
				err := substituteStruct(field.Elem().Addr().Interface(), substitutions)
				if err != nil {
					return err
				}
			}
		}

		// If the field is a struct, perform substitution on its fields.
		if field.Kind() == reflect.Struct {
			err := substituteStruct(field.Addr().Interface(), substitutions)
			if err != nil {
				return err
			}
		}

		// If the field is a map, perform substitution on its values.
		if field.Kind() == reflect.Map {
			for _, key := range field.MapKeys() {
				val := field.MapIndex(key)
				if val.Kind() == reflect.String {
					substitutedVal, err := substitute(val.String(), substitutions)
					if err != nil {
						return err
					}

					field.SetMapIndex(key, reflect.ValueOf(substitutedVal))
				}
			}
		}
	}

	return nil
}

// substitute performs string substitution on the input string
// using the substitutions map.
func substitute(input string, substitutions map[string]string) (string, error) {
	// missing variables are not allowed, Execute will throw an error.
	tmpl, err := template.New("-").Option("missingkey=error").Parse(input)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	err = tmpl.Execute(&result, &substitutions)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

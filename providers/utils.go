package providers

import (
	"errors"
	"html/template"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
)

var ErrProviderCatalogNotFound = errors.New("provider or provider catalog not found")

// ReadConfig reads the configuration from the catalog for specific provider. It also performs string substitution
// on the values in the config that are surrounded by {{}}.
func ReadConfig(provider Provider, substitutions *map[string]string) (*ProviderInfo, error) {
	providerConfig, ok := Catalog[provider]
	if !ok {
		return nil, ErrProviderCatalogNotFound
	}

	// Validate the provider configuration
	v := validator.New()
	if err := v.Struct(&providerConfig); err != nil {
		return nil, err
	}

	// Apply substitutions to the provider configuration values which contain variables in the form of {{var}}.
	err := substituteStruct(&providerConfig, substitutions)
	if err != nil {
		return nil, err
	}

	return &providerConfig, nil
}

// substituteStruct performs string substitution on the fields of the input struct
// using the substitutions map.
func substituteStruct(input interface{}, substitutions *map[string]string) (err error) {
	configStruct := reflect.ValueOf(input).Elem()
	for i := 0; i < configStruct.NumField(); i++ {
		field := configStruct.Field(i)

		// If the field is a string, perform substitution on it.
		if field.Kind() == reflect.String {
			substitutedVal, err := substitute(field.String(), substitutions)
			if err != nil {
				return err
			}

			field.SetString(substitutedVal)
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
func substitute(input string, substitutions *map[string]string) (string, error) {
	tmpl, err := template.New("-").Parse(input)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	err = tmpl.Execute(&result, substitutions)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (i *ProviderInfo) GetOption(key string) (string, bool) {
	if i.ProviderOpts == nil {
		return "", false
	}

	val, ok := i.ProviderOpts[key]

	return val, ok
}

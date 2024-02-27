package providers

import (
	"bytes"
	"encoding/gob"
	"errors"
	"html/template"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
)

var (
	ErrProviderCatalogNotFound = errors.New("provider or provider catalog not found")
	ErrProviderOptionNotFound  = errors.New("provider option not found")
)

func ReadCatalog() (CatalogType, error) {
	catalog, err := clone[CatalogType](catalog)
	if err != nil {
		return nil, err
	}

	// Validate the provider configuration
	v := validator.New()
	for _, providerInfo := range catalog {
		if err := v.Struct(providerInfo); err != nil {
			return nil, err
		}
	}

	return catalog, nil
}

// ReadInfo reads the information from the catalog for specific provider. It also performs string substitution
// on the values in the config that are surrounded by {{}}.
func ReadInfo(provider Provider, substitutions *map[string]string) (*ProviderInfo, error) {
	pInfo, ok := catalog[provider]
	if !ok {
		return nil, ErrProviderCatalogNotFound
	}

	// Clone before modifying
	providerInfo, err := clone[ProviderInfo](pInfo)
	if err != nil {
		return nil, err
	}

	// Validate the provider configuration
	v := validator.New()
	if err := v.Struct(providerInfo); err != nil {
		return nil, err
	}

	if substitutions == nil {
		substitutions = &map[string]string{}
	}

	// Apply substitutions to the provider configuration values which contain variables in the form of {{var}}.
	if err := substituteStruct(&providerInfo, substitutions); err != nil {
		return nil, err
	}

	return &providerInfo, nil
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

// clone uses gob to deep copy objects.
func clone[T any](input T) (T, error) { // nolint:ireturn
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(input); err != nil {
		return input, err
	}

	var clone T
	if err := dec.Decode(&clone); err != nil {
		return input, err
	}

	return clone, nil
}

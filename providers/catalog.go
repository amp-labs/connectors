package providers

import (
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// catalogFileRelativeLoc is the name of the providers.yaml catalog.
	catalogFileRelativeLoc = "providers.yaml"
)

var (
	ErrProviderCatalogNotFound = errors.New("provider catalog file not found")
	ErrUnableToGetCallerCWD    = errors.New("unable to get caller's current working directory")
)

// ReadConfig reads the configuration from the config file for a specific provider. It also performs string substitution
// on the values in the config that are surrounded by {{}}. The provider YAML has more details on how it works.
func ReadConfig(provider Provider, substitutions *map[string]string) (*ProviderInfo, error) {
	config, err := GetCatalog()
	if err != nil {
		return nil, err
	}

	providerConfig, ok := config.Providers[provider]
	if !ok {
		return nil, ErrProviderCatalogNotFound
	}

	// Apply substitutions to the provider configuration values which contain variables in the form of {{var}}.
	err = substituteStruct(&providerConfig, substitutions)
	if err != nil {
		return nil, err
	}

	return &providerConfig, nil
}

// GetCatalog reads the entire provider catalog.
func GetCatalog() (*Catalog, error) {
	// Figure out the cwd of the caller
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, ErrUnableToGetCallerCWD
	}

	// Get the absolute directory of the catalog file
	catalogDir := filepath.Dir(filename)

	// Construct the absolute path to the providers.yaml file
	yamlPath := filepath.Join(catalogDir, catalogFileRelativeLoc)

	// Read the file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}

	catalog := &Catalog{}

	err = yaml.Unmarshal(data, catalog)
	if err != nil {
		return nil, err
	}

	return catalog, nil
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

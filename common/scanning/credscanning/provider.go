// Package credscanning is a wrapper for scanning package.
// Its focus is on scanning Credentials required for catalog provider.
package credscanning

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/providers"
)

var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrProviderInfo     = errors.New("provider info is not understood")
)

// ProviderCredentials is a collection of values for a provider that come either from JSON or ENV.
type ProviderCredentials struct {
	Registry       scanning.Registry
	ProviderValues map[string]string
}

// NewJSONProviderCredentials reads JSON fields that must be present for a provider.
// It performs validation and will tell you fields that are expected for this provider in JSON file.
//
// Note: As of right now there is no way to infer if access token must be provided.
// Therefore, explicitly state via arguments.
func NewJSONProviderCredentials(
	filePath string,
	withRequiredAccessToken bool,
	withRequiredWorkspace bool,
) (*ProviderCredentials, error) {
	return createProviderCreds(getProviderName(filePath), withRequiredAccessToken, withRequiredWorkspace, filePath)
}

// NewENVProviderCredentials reads ENV variables associated with a provider.
func NewENVProviderCredentials(
	providerName string,
	withRequiredAccessToken bool,
	withRequiredWorkspace bool,
) (*ProviderCredentials, error) {
	return createProviderCreds(providerName, withRequiredAccessToken, withRequiredWorkspace, "")
}

func createProviderCreds(
	providerName string, withRequiredAccessToken, withRequiredWorkspace bool, filePath string,
) (*ProviderCredentials, error) {
	// load provider from catalog to imply fields in JSON or ENV vars
	catalog, err := providers.ReadCatalog()
	if err != nil {
		return nil, errors.Join(err, ErrProviderNotFound)
	}

	info, ok := catalog[providerName]
	if !ok {
		return nil, ErrProviderNotFound
	}

	fields, err := getFields(info, withRequiredAccessToken, withRequiredWorkspace)
	if err != nil {
		return nil, err
	}

	registry := scanning.NewRegistry()
	readers := handy.Lists[scanning.Reader]{}

	// add readers for every field
	for kind, fieldList := range fields {
		for _, field := range fieldList {
			reader := selectReader(field, filePath, providerName)
			readers.Add(kind, reader)

			if err = registry.AddReader(reader); err != nil {
				return nil, err
			}
		}
	}

	r := &ProviderCredentials{
		Registry:       registry,
		ProviderValues: make(map[string]string),
	}

	return r, r.loadValues(readers)
}

func selectReader(field Field, filePath string, providerName string) scanning.Reader { // nolint:ireturn
	if len(filePath) != 0 {
		return field.GetJSONReader(filePath) // nolint:ireturn
	}

	return field.GetENVReader(providerName)
}

func (r ProviderCredentials) loadValues(readers handy.Lists[scanning.Reader]) error {
	// validate JSON file or ENV has all Required variables
	missingKeys := make([]string, 0)

	for _, reader := range readers["required"] {
		key, err := reader.Key()
		if err != nil {
			return err
		}

		if value, err := r.Registry.GetString(key); err != nil {
			missingKeys = append(missingKeys, key)
		} else {
			r.ProviderValues[key] = value
		}
	}

	if len(missingKeys) != 0 {
		return fmt.Errorf("%w: %s", scanning.ErrKeyNotFound, strings.Join(missingKeys, ","))
	}

	// Optional values should be saved too
	for _, reader := range readers["optional"] {
		key, err := reader.Key()
		if err != nil {
			return err
		}

		if value, err := r.Registry.GetString(key); err == nil {
			r.ProviderValues[key] = value
		}
	}

	return nil
}

func (r ProviderCredentials) Get(field Field) string {
	return r.ProviderValues[field.Name]
}

func getProviderName(filePath string) string {
	registry := scanning.NewRegistry()

	reader := Fields.Provider.GetJSONReader(filePath)

	err := registry.AddReader(reader)
	if err != nil {
		return ""
	}

	name, _ := registry.GetString(Fields.Provider.Name)

	return name
}

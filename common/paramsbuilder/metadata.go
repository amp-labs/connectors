package paramsbuilder

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

var (
	ErrMissingMetadata             = errors.New("missing authentication metadata")
	ErrIncorrectMetadataParamUsage = errors.New("metadata parameter must have required fields")
)

// Metadata params sets metadata describing authentication information.
type Metadata struct {
	// Map is a registry of metadata values that are needed for connector to function.
	Map map[string]string
	// Connector implementation makes a decision on what fields must be supplied in metadata map by the user.
	// Any missing or empty fields will result into error constructing a connector.
	requiredKeys []string
}

func (m *Metadata) ValidateParams() error {
	if len(m.requiredKeys) == 0 {
		// if metadata is used as a parameter it must have at least one required key.
		return ErrIncorrectMetadataParamUsage
	}

	for _, key := range m.requiredKeys {
		value, ok := m.Map[key]
		if !ok {
			return fmt.Errorf("%w, missing key: %v", ErrMissingMetadata, key)
		}

		if len(value) == 0 {
			return fmt.Errorf("%w, empty key: %v", ErrMissingMetadata, key)
		}
	}

	return nil
}

func (m *Metadata) WithMetadata(metadata map[string]string, requiredKeys []string) {
	m.Map = metadata
	m.requiredKeys = requiredKeys
}

func (m *Metadata) GetCatalogVars() []catalogreplacer.CustomCatalogVariable {
	result := make([]catalogreplacer.CustomCatalogVariable, 0)

	for key, value := range m.Map {
		result = append(result, catalogreplacer.CustomCatalogVariable{
			Plan: catalogreplacer.SubstitutionPlan{
				From: key,
				To:   value,
			},
		})
	}

	return result
}

func (m *Metadata) Value(key string) string {
	return m.Map[key]
}

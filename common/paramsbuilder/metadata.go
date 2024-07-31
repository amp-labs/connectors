package paramsbuilder

import (
	"errors"
	"fmt"
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

func (p *Metadata) ValidateParams() error {
	if len(p.requiredKeys) == 0 {
		// if metadata is used as a parameter it must have at least one required key.
		return ErrIncorrectMetadataParamUsage
	}

	for _, key := range p.requiredKeys {
		value, ok := p.Map[key]
		if !ok {
			return fmt.Errorf("%w, missing key: %v", ErrMissingMetadata, key)
		}

		if len(value) == 0 {
			return fmt.Errorf("%w, empty key: %v", ErrMissingMetadata, key)
		}
	}

	return nil
}

func (p *Metadata) WithMetadata(metadata map[string]string, requiredKeys []string) {
	p.Map = metadata
	p.requiredKeys = requiredKeys
}

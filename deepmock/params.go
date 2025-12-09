package deepmock

import (
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
)

// parameters holds the configuration for the deepmock connector.
type parameters struct {
	paramsbuilder.AuthClient

	rawSchemas    map[string][]byte
	structSchemas map[string]any
	schemas       map[string]*InputSchema

	err error

	observers []func(action string, record map[string]any)
}

// ValidateParams checks that all required parameters are present and valid.
// Note: This method is not part of the NewConnector construction path.
// NewConnector initializes schemas and storage directly and does not call ValidateParams.
// This method should only be used in contexts where schemas and storage are already initialized.
func (p parameters) ValidateParams() error {
	if p.err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSchema, p.err)
	}

	if err := p.AuthClient.ValidateParams(); err != nil {
		return err
	}

	return nil
}

// Option is a function that configures the connector parameters.
type Option = func(*parameters)

// WithObserver adds an observer for modifications to the schema.
func WithObserver(f func(action string, record map[string]any)) Option {
	return func(p *parameters) {
		p.observers = append(p.observers, f)
	}
}

// WithClient wraps an HTTP client in a JSONHTTPClient.
func WithClient(client *http.Client) Option {
	return func(params *parameters) {
		params.AuthClient.WithAuthenticatedClient(client)
	}
}

// WithAuthenticatedClient wraps an authenticated HTTP client.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.AuthClient.WithAuthenticatedClient(client)
	}
}

// WithRawSchemas configures the connector with raw JSON schema definitions.
// The schemas are provided as byte slices that will be parsed during connector initialization.
// These raw schemas take priority over struct-derived schemas if both are provided.
func WithRawSchemas(schemas map[string][]byte) Option {
	return func(p *parameters) {
		p.rawSchemas = schemas
	}
}

// WithSchemas configures the connector with pre-parsed JSON schemas.
// Use this option when you have already parsed jsonschema.Schema objects.
// This is an alternative to providing raw JSON schemas via WithRawSchemas.
func WithSchemas(schemas map[string]*InputSchema) Option {
	return func(p *parameters) {
		p.schemas = schemas
	}
}

// WithStructSchemas configures the connector to derive schemas from Go structs.
// This is an alternative to providing raw JSON schemas in NewConnector.
// If both raw schemas and struct schemas are provided, raw schemas take priority.
func WithStructSchemas(schemas map[string]any) Option {
	return func(params *parameters) {
		params.structSchemas = schemas
	}
}

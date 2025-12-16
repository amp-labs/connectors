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

	// rawSchemas holds unparsed JSON schema definitions as byte slices.
	rawSchemas map[string][]byte
	// structSchemas holds Go struct templates for schema derivation.
	structSchemas map[string]any
	// schemas holds parsed and validated JSON schemas.
	schemas map[string]*InputSchema

	// err holds any error encountered during parameter initialization.
	err error

	// observers holds callback functions that are notified of data modifications.
	observers map[string]func(action string, record map[string]any, metadata map[string]any)

	// storage holds the configured storage backend instance.
	storage Storage

	// storageFactory creates a new storage backend with the given configuration.
	storageFactory func(
		schemas SchemaRegistry,
		idFields, updatedFields map[string]string,
	) (Storage, error)
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
func WithObserver(id string, f func(action string, record map[string]any)) Option {
	return func(p *parameters) {
		p.observers[id] = f
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

// WithStorage configures the connector with a pre-initialized storage backend.
// Use this option when you want to provide a custom storage implementation.
func WithStorage(storage Storage) Option {
	return func(p *parameters) {
		p.storage = storage
	}
}

// WithStorageFactory configures the connector with a factory function for creating storage backends.
// The factory function receives schema information and observers, allowing for dynamic storage initialization.
// This is useful when storage needs to be created with specific runtime configuration.
func WithStorageFactory(f func(
	schemas SchemaRegistry,
	idFields, updatedFields map[string]string,
	observers map[string]func(action string, record map[string]any),
) (Storage, error),
) Option {
	return func(p *parameters) {
		p.storageFactory = f
	}
}

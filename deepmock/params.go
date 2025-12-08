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
	structSchemas map[string]interface{}
	err           error
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

// WithStructSchemas configures the connector to derive schemas from Go structs.
// This is an alternative to providing raw JSON schemas in NewConnector.
// If both raw schemas and struct schemas are provided, raw schemas take priority.
func WithStructSchemas(schemas map[string]interface{}) Option {
	return func(params *parameters) {
		params.structSchemas = schemas
	}
}

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
	schemas schemaRegistry
	storage *Storage
	err     error
}

// ValidateParams checks that all required parameters are present and valid.
func (p parameters) ValidateParams() error {
	if p.err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSchema, p.err)
	}

	if err := p.AuthClient.ValidateParams(); err != nil {
		return err
	}

	if p.schemas == nil || len(p.schemas) == 0 {
		return fmt.Errorf("%w: schemas", ErrMissingParam)
	}

	if p.storage == nil {
		return fmt.Errorf("%w: storage", ErrMissingParam)
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

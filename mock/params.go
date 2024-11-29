package mock

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// Option is a function which mutates the hubspot connector configuration.
type Option = func(params *parameters)

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(client *http.Client) Option {
	return func(params *parameters) {
		WithAuthenticatedClient(client)(params)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: client,
			},
		}
	}
}

// WithRead sets the read function for the connector.
func WithRead(read func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)) Option {
	return func(params *parameters) {
		params.read = read
	}
}

// WithWrite sets the write function for the connector.
func WithWrite(write func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)) Option {
	return func(params *parameters) {
		params.write = write
	}
}

// WithListObjectMetadata sets the listObjectMetadata function for the connector.
func WithListObjectMetadata(
	listObjectMetadata func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error),
) Option {
	return func(params *parameters) {
		params.listObjectMetadata = listObjectMetadata
	}
}

// parameters is the internal configuration for the mock connector.
type parameters struct {
	client             *common.JSONHTTPClient // required
	read               func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)
	write              func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)
	listObjectMetadata func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error)
}

func (p parameters) ValidateParams() error {
	if p.client == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "client")
	}

	if p.read == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "read")
	}

	if p.write == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "write")
	}

	if p.listObjectMetadata == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "listObjectMetadata")
	}

	return nil
}

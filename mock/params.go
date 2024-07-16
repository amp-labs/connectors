package mock

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// Option is a function which mutates the hubspot connector configuration.
type Option func(params *mockParams)

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(client *http.Client) Option {
	return func(params *mockParams) {
		WithAuthenticatedClient(client)(params)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *mockParams) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: client,
			},
		}
	}
}

// WithRead sets the read function for the connector.
func WithRead(read func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)) Option {
	return func(params *mockParams) {
		params.read = read
	}
}

// WithWrite sets the write function for the connector.
func WithWrite(write func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)) Option {
	return func(params *mockParams) {
		params.write = write
	}
}

// WithListObjectMetadata sets the listObjectMetadata function for the connector.
func WithListObjectMetadata(
	listObjectMetadata func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error),
) Option {
	return func(params *mockParams) {
		params.listObjectMetadata = listObjectMetadata
	}
}

// mockParams is the internal configuration for the mock connector.
type mockParams struct {
	client             *common.JSONHTTPClient // required
	read               func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)
	write              func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)
	listObjectMetadata func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error)
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *mockParams) prepare() (out *mockParams, err error) {
	if p.client == nil {
		return nil, fmt.Errorf("%w: %s", ErrMissingParam, "client")
	}

	if p.read == nil {
		return nil, fmt.Errorf("%w: %s", ErrMissingParam, "read")
	}

	if p.write == nil {
		return nil, fmt.Errorf("%w: %s", ErrMissingParam, "write")
	}

	if p.listObjectMetadata == nil {
		return nil, fmt.Errorf("%w: %s", ErrMissingParam, "listObjectMetadata")
	}

	return p, nil
}

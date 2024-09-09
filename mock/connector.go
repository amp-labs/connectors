package mock

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	client *common.JSONHTTPClient

	read               func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)
	write              func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)
	listObjectMetadata func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error)
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params := &mockParams{
		read: func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "read")
		},
		write: func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "write")
		},
		listObjectMetadata: func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "listObjectMetadata")
		},
	}
	for _, opt := range opts {
		opt(params)
	}

	params, err := params.prepare()
	if err != nil {
		return nil, err
	}

	return &Connector{
		client:             params.client,
		read:               params.read,
		write:              params.write,
		listObjectMetadata: params.listObjectMetadata,
	}, nil
}

func (c *Connector) String() string {
	return "mock"
}

func (c *Connector) Close() error {
	return nil
}

func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.client
}

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.client.HTTPClient
}

func (c *Connector) Provider() providers.Provider {
	return providers.Mock
}

func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	return c.read(ctx, params)
}

func (c *Connector) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	return c.write(ctx, params)
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	return c.listObjectMetadata(ctx, objectNames)
}

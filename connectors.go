package connectors

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/hubspot"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/proxy"
	"github.com/amp-labs/connectors/salesforce"
)

// ProxyConnector is an interface that can be used to implement a connector with
// that allows proxying requests through to a provider.
type ProxyConnector interface {
	fmt.Stringer

	// Provider returns the connector's provider.
	Provider() providers.Provider

	// ProviderInfo returns the connector's provider.
	ProviderInfo() *providers.ProviderInfo

	// JSONClient returns the connector's JSON client.
	JSONClient() *common.JSONHTTPClient

	// HTTPClient returns the connector's HTTP client.
	HTTPClient() *common.HTTPClient
}

// ReadConnector is an interface that can be used to implement a connector that
// can read data from a provider.
type ReadConnector interface {
	ProxyConnector

	// Read reads a page of data from the connector. This can be called multiple
	// times to read all the data. The caller is responsible for paging, by
	// passing the NextPage value correctly, and by terminating the loop when
	// Done is true. The caller is also responsible for handling errors.
	// Authentication corner cases are handled internally, but all other errors
	// are returned to the caller.
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)

	ListObjectMetadata(ctx context.Context, objectNames []string) (*ListObjectMetadataResult, error)
}

// WriteConnector is an interface that can be used to implement a connector
// that can write data to a provider.
type WriteConnector interface {
	ProxyConnector

	Write(ctx context.Context, params WriteParams) (*WriteResult, error)
}

// We re-export the following types so that they can be used by consumers of this library.
type (
	ReadParams               = common.ReadParams
	WriteParams              = common.WriteParams
	ReadResult               = common.ReadResult
	WriteResult              = common.WriteResult
	ListObjectMetadataResult = common.ListObjectMetadataResult

	ErrorWithStatus = common.HTTPStatusError
)

// We re-export the following errors so that they can be handled by consumers of this library.
var (
	// ErrAccessToken represents a token which isn't valid.
	ErrAccessToken = common.ErrAccessToken

	// ErrApiDisabled means a customer didn't enable this API on their SaaS instance.
	ErrApiDisabled = common.ErrApiDisabled

	// ErrRetryable represents a temporary error. Can retry.
	ErrRetryable = common.ErrRetryable

	// ErrCaller represents non-retryable errors caused by bad input from the caller.
	ErrCaller = common.ErrCaller

	// ErrServer represents non-retryable errors caused by something on the server.
	ErrServer = common.ErrServer

	// ErrUnknownConnector represents an unknown connector.
	ErrUnknownConnector = errors.New("unknown connector")
)

// NewProxyConnector returns a new proxy connector for the given provider with the given build options.
func NewProxyConnector(
	provider providers.Provider,
	opts ...proxy.Option,
) (*proxy.Connector, error) {
	return proxy.NewConnector(provider, opts...)
}

// NewReadConnector returns a new read connector for the given provider with the given build options.
func NewReadConnector(
	provider providers.Provider,
	opts ...proxy.Option,
) (ReadConnector, error) { // nolint:ireturn
	var (
		conn   ReadConnector
		outErr error
	)

	proxyConn, err := NewProxyConnector(provider, opts...)
	if err != nil {
		return nil, err
	}

	switch provider {
	case providers.Hubspot:
		conn, outErr = hubspot.NewConnector(hubspot.WithProxyConnector(proxyConn))
	case providers.Salesforce:
		conn, outErr = salesforce.NewConnector(salesforce.WithProxyConnector(proxyConn))
	default:
		conn = nil
		outErr = ErrUnknownConnector
	}

	return conn, outErr
}

// NewWriteConnector returns a new write connector for the given provider with the given build options.
func NewWriteConnector(
	provider providers.Provider,
	opts ...proxy.Option,
) (WriteConnector, error) { // nolint:ireturn
	var (
		conn   WriteConnector
		outErr error
	)

	proxyConn, err := NewProxyConnector(provider, opts...)
	if err != nil {
		return nil, err
	}

	switch provider {
	case providers.Hubspot:
		conn, outErr = hubspot.NewConnector(hubspot.WithProxyConnector(proxyConn))
	case providers.Salesforce:
		conn, outErr = salesforce.NewConnector(salesforce.WithProxyConnector(proxyConn))
	default:
		conn = nil
		outErr = ErrUnknownConnector
	}

	return conn, outErr
}

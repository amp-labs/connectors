package connectors

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/hubspot"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/salesforce"
	"github.com/amp-labs/connectors/zendesk"
)

// BasicConnector is an interface that can be used to implement a connector with
// basic configuration about the provider.
type BasicConnector interface {
	fmt.Stringer
	io.Closer

	// HTTPClient returns the underlying HTTP client. This is useful for proxy requests.
	HTTPClient() *common.HTTPClient

	// Provider returns the connector provider.
	Provider() providers.Provider
}

// Connector is an interface that all connectors must implement.
type Connector interface {
	BasicConnector

	// Read reads a page of data from the connector. This can be called multiple
	// times to read all the data. The caller is responsible for paging, by
	// passing the NextPage value correctly, and by terminating the loop when
	// Done is true. The caller is also responsible for handling errors.
	// Authentication corner cases are handled internally, but all other errors
	// are returned to the caller.
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)

	Write(ctx context.Context, params WriteParams) (*WriteResult, error)

	ListObjectMetadata(ctx context.Context, objectNames []string) (*ListObjectMetadataResult, error)

	// JSONHTTPClient returns the underlying JSON HTTP client. This is useful for
	// testing, or for calling methods that aren't exposed by the Connector
	// interface directly. Authentication and token refreshes will be handled automatically.
	JSONHTTPClient() *common.JSONHTTPClient
}

// API is a function that returns a Connector. It's used as a factory.
type API[Conn Connector, Option any] func(opts ...Option) (Conn, error)

// New returns a new Connector. It's a convenience wrapper around the API.
func (a API[Conn, Option]) New(opts ...Option) (Connector, error) { //nolint:ireturn
	if a == nil {
		return nil, ErrUnknownConnector
	}

	return a(opts...)
}

// Salesforce is an API that returns a new Salesforce Connector.
var Salesforce API[*salesforce.Connector, salesforce.Option] = salesforce.NewConnector //nolint:gochecknoglobals

// Hubspot is an API that returns a new Hubspot Connector.
var Hubspot API[*hubspot.Connector, hubspot.Option] = hubspot.NewConnector //nolint:gochecknoglobals

// Zendesk is an API that returns a new Zendesk Connector.
var Zendesk API[*zendesk.Connector, zendesk.Option] = zendesk.NewConnector //nolint:gochecknoglobals

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

// New returns a new Connector. The signature is generic to facilitate more flexible caller setup
// (e.g. constructing a new connector based on parsing a config file, whose exact params
// aren't known until runtime). However, if you can use the API.New form, it's preferred,
// since you get type safety and more readable code.
func New(provider providers.Provider, opts map[string]any) (Connector, error) { //nolint:ireturn
	switch provider {
	case providers.Salesforce:
		return newSalesforce(opts)
	case providers.Hubspot:
		return newHubspot(opts)
	case providers.Zendesk:
		return newZendesk(opts)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownConnector, provider)
	}
}

// newSalesforce returns a new Salesforce Connector, by unwrapping the options and passing them to the Salesforce API.
func newSalesforce(opts map[string]any) (Connector, error) { //nolint:ireturn
	var options []salesforce.Option

	c, valid := getParam[common.AuthenticatedHTTPClient](opts, "client")
	if valid {
		options = append(options, salesforce.WithAuthenticatedClient(c))
	}

	w, valid := getParam[string](opts, "workspace")
	if valid {
		options = append(options, salesforce.WithWorkspace(w))
	}

	return Salesforce.New(options...)
}

// newHubspot returns a new Hubspot Connector, by unwrapping the options and passing them to the Hubspot API.
func newHubspot(opts map[string]any) (Connector, error) { //nolint:ireturn
	var options []hubspot.Option

	c, valid := getParam[common.AuthenticatedHTTPClient](opts, "client")
	if valid {
		options = append(options, hubspot.WithAuthenticatedClient(c))
	}

	w, valid := getParam[hubspot.APIModule](opts, "module")
	if valid {
		options = append(options, hubspot.WithModule(w))
	}

	return Hubspot.New(options...)
}

// newZendesk returns a new Zendesk Connector, by unwrapping the options and passing them to the Zendesk API.
func newZendesk(opts map[string]any) (Connector, error) { //nolint:ireturn
	var options []zendesk.Option

	c, valid := getParam[common.AuthenticatedHTTPClient](opts, "client")
	if valid {
		options = append(options, zendesk.WithAuthenticatedClient(c))
	}

	w, valid := getParam[string](opts, "workspace")
	if valid {
		options = append(options, zendesk.WithWorkspace(w))
	}

	return Zendesk.New(options...)
}

// getParam returns the value of the given key, if present, safely cast to an assumed type.
// If the key is not present, or the value is not of the assumed type, it returns the
// zero value of the desired type, and false. In case of success, it returns the value and true.
func getParam[A any](opts map[string]any, key string) (A, bool) { //nolint:ireturn
	var zero A

	if opts == nil {
		return zero, false
	}

	val, present := opts[key]
	if !present {
		return zero, false
	}

	a, ok := val.(A)
	if !ok {
		return zero, false
	}

	return a, true
}

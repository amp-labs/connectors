package connectors

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/dynamicscrm"
	"github.com/amp-labs/connectors/hubspot"
	"github.com/amp-labs/connectors/intercom"
	"github.com/amp-labs/connectors/mock"
	"github.com/amp-labs/connectors/outreach"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/salesforce"
	"github.com/amp-labs/connectors/salesloft"
)

// Connector is an interface that can be used to implement a connector with
// basic configuration about the provider.
type Connector interface {
	fmt.Stringer
	io.Closer

	// JSONHTTPClient returns the underlying JSON HTTP client. This is useful for
	// testing, or for calling methods that aren't exposed by the Connector
	// interface directly. Authentication and token refreshes will be handled automatically.
	JSONHTTPClient() *common.JSONHTTPClient

	// HTTPClient returns the underlying HTTP client. This is useful for proxy requests.
	HTTPClient() *common.HTTPClient

	// Provider returns the connector provider.
	Provider() providers.Provider
}

// ReadConnector is an interface that extends the Connector interface with read capabilities.
type ReadConnector interface {
	Connector

	// Read reads a page of data from the connector. This can be called multiple
	// times to read all the data. The caller is responsible for paging, by
	// passing the NextPage value correctly, and by terminating the loop when
	// Done is true. The caller is also responsible for handling errors.
	// Authentication corner cases are handled internally, but all other errors
	// are returned to the caller.
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)
}

// WriteConnector is an interface that extends the Connector interface with write capabilities.
type WriteConnector interface {
	Connector

	Write(ctx context.Context, params WriteParams) (*WriteResult, error)
}

// DeleteConnector is an interface that extends the Connector interface with delete capabilities.
type DeleteConnector interface {
	Connector

	Delete(ctx context.Context, params DeleteParams) (*DeleteResult, error)
}

// ObjectMetadataConnector is an interface that extends the Connector interface with
// the ability to list object metadata.
type ObjectMetadataConnector interface {
	Connector

	ListObjectMetadata(ctx context.Context, objectNames []string) (*ListObjectMetadataResult, error)
}

type AuthMetadataConnector interface {
	Connector

	GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error)
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

// DynamicsCRM is an API that returns a new Microsoft Dynamics 365 CRM Connector.
var DynamicsCRM API[*dynamicscrm.Connector, dynamicscrm.Option] = dynamicscrm.NewConnector //nolint:gochecknoglobals,lll

// Mock is an API that returns a new Mock Connector.
var Mock API[*mock.Connector, mock.Option] = mock.NewConnector //nolint:gochecknoglobals

// Outreach is an API that returns a new Outreach Connector.
var Outreach API[*outreach.Connector, outreach.Option] = outreach.NewConnector //nolint:gochecknoglobals

// Salesloft is an API that returns a new Salesloft Connector.
var Salesloft API[*salesloft.Connector, salesloft.Option] = salesloft.NewConnector //nolint:gochecknoglobals

// Intercom is an API that returns a new Intercom Connector.
var Intercom API[*intercom.Connector, intercom.Option] = intercom.NewConnector //nolint:gochecknoglobals

// We re-export the following types so that they can be used by consumers of this library.
type (
	ReadParams               = common.ReadParams
	WriteParams              = common.WriteParams
	DeleteParams             = common.DeleteParams
	ReadResult               = common.ReadResult
	WriteResult              = common.WriteResult
	DeleteResult             = common.DeleteResult
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
	case providers.Mock:
		return newMock(opts)
	case providers.Salesforce:
		return newSalesforce(opts)
	case providers.Hubspot:
		return newHubspot(opts)
	case providers.Outreach:
		return newOutreach(opts)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownConnector, provider)
	}
}

// newMock returns a new Mock Connector, by unwrapping the options and passing them to the Mock API.
func newMock(opts map[string]any) (Connector, error) { //nolint:ireturn
	var options []mock.Option

	c, valid := getParam[*http.Client](opts, "client")
	if valid {
		options = append(options, mock.WithClient(c))
	}

	a, valid := getParam[common.AuthenticatedHTTPClient](opts, "authenticated-client")
	if valid {
		options = append(options, mock.WithAuthenticatedClient(a))
	}

	r, valid := getParam[func(ctx context.Context, params ReadParams) (*ReadResult, error)](opts, "read")
	if valid {
		options = append(options, mock.WithRead(r))
	}

	w, valid := getParam[func(ctx context.Context, params WriteParams) (*WriteResult, error)](opts, "write")
	if valid {
		options = append(options, mock.WithWrite(w))
	}

	l, valid := getParam[func(ctx context.Context, objectNames []string) (*ListObjectMetadataResult, error)](
		opts, "list-object-metadata")
	if valid {
		options = append(options, mock.WithListObjectMetadata(l))
	}

	return Mock.New(options...)
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

// newOutreach returns a new Outreach Connector, by unwrapping the options and passing them to the Outreach API.
func newOutreach(opts map[string]any) (Connector, error) { //nolint:ireturn
	var options []outreach.Option

	c, valid := getParam[common.AuthenticatedHTTPClient](opts, "client")
	if valid {
		options = append(options, outreach.WithAuthenticatedClient(c))
	}

	return Outreach.New(options...)
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

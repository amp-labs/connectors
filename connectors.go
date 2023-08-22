package connectors

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/salesforce"
)

// TODO: Make a transport type that can be used to configure the http.Client.

// Connector is an interface that all connectors must implement.
type Connector interface {
	fmt.Stringer
	io.Closer

	Name() string
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)

	HTTPClient() HTTPClient
}

// API is a function that returns a Connector. It's used as a factory.
type API[Conn Connector, Option any] func(opts ...Option) (Conn, error)

// Salesforce is an API that returns a new Salesforce Connector.
var Salesforce API[*salesforce.Connector, salesforce.Option] = salesforce.NewConnector //nolint:gochecknoglobals

// We re-export the following types so that they can be used by consumers of this library.
type (
	ReadParams      = common.ReadParams
	ReadResult      = common.ReadResult
	ErrorWithStatus = common.HTTPStatusError
	HTTPClient      = common.HTTPClient
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

	// ErrUnknownConnector represents an unknown connector name.
	ErrUnknownConnector = errors.New("unknown connector")
)

// New returns a new Connector.
func New[Conn Connector, Option any](api API[Conn, Option], //nolint:ireturn
	opts ...Option,
) (Connector, error) {
	return api(opts...)
}

func Providers() []string {
	return []string{
		salesforce.Name,
	}
}

package connectors

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/salesforce"
)

// Connector is an interface that all connectors must implement.
type Connector interface {
	fmt.Stringer
	io.Closer

	Name() string
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)
}

// API is a function that returns a Connector. It's used as a factory.
type API[Conn Connector, Token any] func(workspaceRef string, getToken common.TokenProvider[Token]) Conn

// Salesforce is an API that returns a new Salesforce Connector.
var Salesforce API[*salesforce.Connector, string] = salesforce.NewConnector //nolint:gochecknoglobals

// We re-export the following types so that they can be used by consumers of this library.
type (
	ReadParams      = common.ReadParams
	ReadResult      = common.ReadResult
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

	ErrUnknownConnector = errors.New("unknown connector")
)

// New returns a new Connector.
func New[Conn Connector, Token any](api API[Conn, Token], workspaceRef string, //nolint:ireturn
	getToken func(ctx context.Context) (Token, error),
) Connector {
	return api(workspaceRef, getToken)
}

func NewFromProviderName(name string, workspaceRef string, //nolint:ireturn
	getToken func(ctx context.Context) (any, error),
) (Connector, error) {
	switch name {
	case salesforce.Name:
		return New(Salesforce, workspaceRef, wrapStringTokenProvider(getToken)), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownConnector, name)
	}
}

func Providers() []string {
	return []string{
		salesforce.Name,
	}
}

func wrapStringTokenProvider(f func(ctx context.Context) (any, error)) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		tok, err := f(ctx)
		if err != nil {
			return "", err
		}

		token, ok := tok.(string)
		if !ok {
			return "", fmt.Errorf("invalid token type: %T", tok) //nolint:goerr113
		}

		return token, nil
	}
}

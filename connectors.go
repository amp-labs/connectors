package connectors

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/salesforce"
	"golang.org/x/oauth2"
)

// Connector is an interface that all connectors must implement.
type Connector interface {
	fmt.Stringer
	io.Closer

	Name() string
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)

	HTTPClient() HTTPClient
}

// API is a function that returns a Connector. It's used as a factory.
type API[Conn Connector, Option any] func(ctx context.Context, opts ...Option) (Conn, error)

func (a API[Conn, Option]) New(ctx context.Context, opts ...Option) (Connector, error) { //nolint:ireturn
	if a == nil {
		return nil, ErrUnknownConnector
	}

	return a(getContext(ctx), opts...)
}

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

	// ErrUnknownConnector represents an unknown connector.
	ErrUnknownConnector = errors.New("unknown connector")
)

// getContext returns a context, or a background context if the given context is nil.
func getContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	} else {
		return ctx
	}
}

// APINames returns a list of supported connector names.
func APINames() []string {
	return []string{
		salesforce.Name,
	}
}

// New returns a new Connector. The signature is generic to facilitate more flexible caller setup
// (e.g. constructing a new connector based on parsing a config file, whose exact params
// aren't known until runtime). However, if you can use the API.New form, it's preferred,
// since you get type safety and more intuitive argument names.
func New(ctx context.Context, apiName string, opts map[string]any) (Connector, error) { //nolint:ireturn
	if strings.EqualFold(apiName, salesforce.Name) {
		return newSalesforce(ctx, opts)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnknownConnector, apiName)
}

func newSalesforce(ctx context.Context, opts map[string]any) (Connector, error) {
	var options []salesforce.Option

	c, ok := opts["client"]
	if ok {
		cl, ok := c.(*http.Client)
		if !ok {
			return nil, fmt.Errorf("invalid client type: %T (expected *http.Client)", c)
		}

		options = append(options, salesforce.WithClient(cl))
	}

	w, ok := opts["workspace"]
	if ok {
		wr, ok := w.(string)
		if !ok {
			return nil, fmt.Errorf("invalid workspace type: %T (expected string)", w)
		}

		options = append(options, salesforce.WithWorkspace(wr))
	}

	ot, ok := opts["oauth_token"]
	if ok {
		otk, ok := ot.(*oauth2.Token)
		if !ok {
			return nil, fmt.Errorf("invalid oauth_token type: %T (expected *oauth2.Token)", ot)
		}

		options = append(options, salesforce.WithOAuthToken(otk))
	}

	oc, ok := opts["oauth_config"]
	if ok {
		ocf, ok := oc.(*oauth2.Config)
		if !ok {
			return nil, fmt.Errorf("invalid oauth_config type: %T (expected *oauth2.Config)", oc)
		}

		options = append(options, salesforce.WithOAuthConfig(ocf))
	}

	ts, ok := opts["token_source"]
	if ok {
		tsk, ok := ts.(oauth2.TokenSource)
		if !ok {
			return nil, fmt.Errorf("invalid token_source type: %T (expected oauth2.TokenSource)", ts)
		}

		options = append(options, salesforce.WithTokenSource(tsk))
	}

	return Salesforce.New(ctx, options...)
}

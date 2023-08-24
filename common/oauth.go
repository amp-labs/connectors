package common

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type AuthenticatedHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

// NewOAuthHTTPClient returns a new http client, with automatic OAuth authentication. Specifically
// this means that the client will automatically refresh the access token whenever it expires.
func NewOAuthHTTPClient(ctx context.Context, opts ...OAuthOption) (AuthenticatedHTTPClient, error) { //nolint:ireturn
	params := &oauthClientParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error
	params, err = params.prepare()

	if err != nil {
		return nil, err
	}

	return newOAuthClient(ctx, params), nil
}

// oauthClientParams is the internal configuration for the oauth http client.
type oauthClientParams struct {
	client      *http.Client
	token       *oauth2.Token
	config      *oauth2.Config
	tokenSource oauth2.TokenSource
}

type OAuthOption func(params *oauthClientParams)

// WithClient sets the http client to use for the connector. Its usage is optional.
func WithClient(client *http.Client) OAuthOption {
	return func(params *oauthClientParams) {
		params.client = client
	}
}

// WithOAuthToken sets the oauth token to use for the connector. It's required,
// unless a token source is provided.
func WithOAuthToken(token *oauth2.Token) OAuthOption {
	return func(params *oauthClientParams) {
		params.token = token
	}
}

// WithOAuthConfig sets the oauth config to use for the connector. It's required,
// unless a token source is provided.
func WithOAuthConfig(config *oauth2.Config) OAuthOption {
	return func(params *oauthClientParams) {
		params.config = config
	}
}

// WithTokenSource sets the oauth token source to use for the connector. Whenever
// the token expires, this will be called to refresh it.
func WithTokenSource(tokenSource oauth2.TokenSource) OAuthOption {
	return func(params *oauthClientParams) {
		params.tokenSource = tokenSource
	}
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *oauthClientParams) prepare() (*oauthClientParams, error) {
	if p.client == nil {
		p.client = http.DefaultClient
	}

	if p.tokenSource == nil {
		if p.token == nil {
			return nil, ErrMissingRefreshToken
		}

		if p.config == nil {
			return nil, ErrMissingOauthConfig
		}
	}

	return p, nil
}

// newHTTPClient returns a new http client for the connector, with automatic OAuth authentication.
func newOAuthClient(ctx context.Context, params *oauthClientParams) AuthenticatedHTTPClient { //nolint:ireturn
	// This is how the key refresher accepts a custom http client
	ctx = context.WithValue(ctx, oauth2.HTTPClient, params.client)

	// Returns a new client which automatically refreshes the access token
	// whenever the current one expires.
	if params.tokenSource != nil {
		return oauth2.NewClient(ctx, params.tokenSource)
	} else {
		return oauth2.NewClient(ctx, params.config.TokenSource(ctx, params.token))
	}
}

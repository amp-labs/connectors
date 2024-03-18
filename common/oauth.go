package common

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

// AuthenticatedHTTPClient is an interface for an http client which can automatically
// authenticate itself. This is useful for OAuth authentication, where the access token
// needs to be refreshed automatically. The signatures are a subset of http.Client,
// so it can be used as a (mostly) drop-in replacement.
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
	client       *http.Client
	token        *oauth2.Token
	config       *oauth2.Config
	tokenSource  oauth2.TokenSource
	tokenUpdated func(oldToken, newToken *oauth2.Token) error
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

// WithTokenUpdated sets the function to call whenever the oauth token is updated.
// This is useful for persisting the refreshed tokens somewhere, so that it can be
// used later. It's optional.
func WithTokenUpdated(onTokenUpdated func(oldToken, newToken *oauth2.Token) error) OAuthOption {
	return func(params *oauthClientParams) {
		params.tokenUpdated = onTokenUpdated
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

	tokenSource := getTokenSource(ctx, params)
	if params.tokenUpdated != nil {
		tokenSource = &observableTokenSource{
			tokenUpdated: params.tokenUpdated,
			lastKnown:    params.token,
			tokenSource:  tokenSource,
		}
	}

	// Returns a new client which automatically refreshes the access token
	// whenever the current one expires.
	return oauth2.NewClient(ctx, tokenSource)
}

func getTokenSource(ctx context.Context, params *oauthClientParams) oauth2.TokenSource { //nolint:ireturn
	if params.tokenSource != nil {
		return params.tokenSource
	}

	return params.config.TokenSource(ctx, params.token)
}

type observableTokenSource struct {
	mut          sync.Mutex
	tokenUpdated func(oldToken, newToken *oauth2.Token) error
	lastKnown    *oauth2.Token
	tokenSource  oauth2.TokenSource
}

func (w *observableTokenSource) Token() (*oauth2.Token, error) {
	w.mut.Lock()
	defer w.mut.Unlock()

	tok, err := w.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	if w.HasChanged(tok) {
		if err := w.tokenUpdated(w.lastKnown, tok); err != nil {
			return nil, err
		}

		w.lastKnown = tok
	}

	return tok, nil
}

func (w *observableTokenSource) HasChanged(tok *oauth2.Token) bool {
	if w.lastKnown == nil {
		return true
	}

	return w.lastKnown.AccessToken == tok.AccessToken ||
		w.lastKnown.RefreshToken == tok.RefreshToken ||
		w.lastKnown.TokenType == tok.TokenType ||
		w.lastKnown.Expiry.Equal(tok.Expiry)
}

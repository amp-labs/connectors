package common

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
)

type OAuthOption func(*oauthClientParams)

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
	debug        func(req *http.Request, rsp *http.Response)
}

// WithOAuthClient sets the http client to use for the connector. Its usage is optional.
func WithOAuthClient(client *http.Client) OAuthOption {
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

// WithOAuthDebug sets a debug function to be called on every request and response,
// after the response has been received from the downstream API.
func WithOAuthDebug(f func(req *http.Request, rsp *http.Response)) OAuthOption {
	return func(params *oauthClientParams) {
		params.debug = f
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
	return &http.Client{
		Transport: &oauth2Transport{
			Source: tokenSource,
			Base:   params.client.Transport,
			Debug:  params.debug,
		},
	}
}

type oauth2Transport struct {
	Source oauth2.TokenSource
	Base   http.RoundTripper
	Debug  func(req *http.Request, rsp *http.Response)
}

func (t *oauth2Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBodyClosed := false

	if req.Body != nil {
		defer func() {
			if !reqBodyClosed {
				_ = req.Body.Close()
			}
		}()
	}

	token, err := t.Source.Token()
	if err != nil {
		return nil, err
	}

	req2 := cloneRequest(req) // per RoundTripper contract
	token.SetAuthHeader(req2)

	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true

	rsp, err := t.base().RoundTrip(req2)
	if err != nil {
		return rsp, err
	}

	if t.Debug != nil {
		t.Debug(req2, cloneResponse(rsp))
	}

	return rsp, nil
}

func (t *oauth2Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}

	return http.DefaultTransport
}

func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r

	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}

	return r2.WithContext(r.Context())
}

func cloneResponse(r *http.Response) *http.Response {
	// shallow copy of the struct
	r2 := new(http.Response)
	*r2 = *r

	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}

	return r2
}

func getTokenSource(ctx context.Context, params *oauthClientParams) oauth2.TokenSource { //nolint:ireturn
	if params.tokenSource != nil {
		return params.tokenSource
	}

	if _, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); !ok {
		if params.client != nil {
			ctx = context.WithValue(ctx, oauth2.HTTPClient, params.client)
		}
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

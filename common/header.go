// nolint:revive,godoclint
package common

import (
	"context"
	"net/http"
)

type HeaderAuthClientOption func(params *headerClientParams)

var WithTrailingSlash = "/" //nolint:gochecknoglobals

// NewHeaderAuthHTTPClient returns a new http client, which will
// do generic header-based authentication. It does this by automatically
// adding the provided headers to every request. There's no additional
// logic for refreshing tokens or anything like that. This is appropriate
// for APIs that use keys or basic auth.
func NewHeaderAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	opts ...HeaderAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	params := &headerClientParams{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(params)
	}

	return newHeaderAuthClient(ctx, params.prepare()), nil
}

// WithHeaderClient sets the http client to use for the connector. Its usage is optional.
func WithHeaderClient(client *http.Client) HeaderAuthClientOption {
	return func(params *headerClientParams) {
		params.client = client
	}
}

// WithHeaderDebug sets a debug function to be called on every request and response,
// after the response has been received from the downstream API.
func WithHeaderDebug(f func(req *http.Request, rsp *http.Response)) HeaderAuthClientOption {
	return func(params *headerClientParams) {
		params.debug = f
	}
}

// WithHeaders sets the headers to use for the connector. Its usage is optional.
func WithHeaders(headers ...Header) HeaderAuthClientOption {
	return func(params *headerClientParams) {
		params.headers = append(params.headers, headers...)
	}
}

// WithHeaderUnauthorizedHandler sets the function to call whenever the response is 401 unauthorized.
// This is useful for handling the case where the server has invalidated the credentials, and the client
// needs to refresh. It's optional.
func WithHeaderUnauthorizedHandler(
	f func(hdrs []Header, req *http.Request, rsp *http.Response) (*http.Response, error),
) HeaderAuthClientOption {
	return func(params *headerClientParams) {
		params.unauthorized = f
	}
}

// WithHeaderIsUnauthorizedHandler sets the function to call
// whenever the response is unauthorized (not necessarily 401).
// This is useful for handling the case where the server has invalidated the token, and the client
// needs to forcefully refresh. It's optional.
func WithHeaderIsUnauthorizedHandler(
	f func(rsp *http.Response) bool,
) HeaderAuthClientOption {
	return func(params *headerClientParams) {
		params.isUnauthorized = f
	}
}

// WithDynamicHeaders sets a function that will be called on every request to
// get additional headers to use. Use this for things like time-based tokens
// or loading headers from some external authority. The function can access a
// copy of the request object to use its metadata for generating headers.
func WithDynamicHeaders(f DynamicHeadersGenerator) HeaderAuthClientOption {
	return func(params *headerClientParams) {
		params.dynamicHeaders = f
	}
}

type DynamicHeadersGenerator func(*http.Request) ([]Header, error)

// oauthClientParams is the internal configuration for the oauth http client.
type headerClientParams struct {
	client         *http.Client
	headers        []Header
	dynamicHeaders DynamicHeadersGenerator
	debug          func(req *http.Request, rsp *http.Response)
	unauthorized   func(hdrs []Header, req *http.Request, rsp *http.Response) (*http.Response, error)
	isUnauthorized func(rsp *http.Response) bool
}

func (p *headerClientParams) prepare() *headerClientParams {
	if p.client == nil {
		p.client = http.DefaultClient
	}

	return p
}

// newHTTPClient returns a new http client for the connector, with automatic OAuth authentication.
func newHeaderAuthClient(_ context.Context, params *headerClientParams) AuthenticatedHTTPClient { //nolint:ireturn
	return &headerAuthClient{
		client:         params.client,
		headers:        params.headers,
		dynamicHeaders: params.dynamicHeaders,
		debug:          params.debug,
		unauthorized:   params.unauthorized,
		isUnauthorized: params.isUnauthorized,
	}
}

type headerAuthClient struct {
	client         *http.Client
	headers        []Header
	dynamicHeaders DynamicHeadersGenerator
	debug          func(req *http.Request, rsp *http.Response)
	unauthorized   func(hdrs []Header, req *http.Request, rsp *http.Response) (*http.Response, error)
	isUnauthorized func(rsp *http.Response) bool
}

func (c *headerAuthClient) Do(req *http.Request) (*http.Response, error) {
	// This allows us to attach headers without modifying the input
	req2 := req.Clone(req.Context())

	for _, header := range c.headers {
		header.ApplyToRequest(req2)
	}

	if c.dynamicHeaders != nil {
		hdrs, err := c.dynamicHeaders(req2)
		if err != nil {
			return nil, err
		}

		for _, header := range hdrs {
			header.ApplyToRequest(req2)
		}
	}

	modifier, hasModifier := getRequestModifier(req2.Context()) //nolint:contextcheck
	if hasModifier {
		modifier(req2)
	}

	rsp, err := c.client.Do(req2)
	if err != nil {
		return rsp, err
	}

	if c.debug != nil {
		c.debug(req2, cloneResponse(rsp))
	}

	return c.handleUnauthorizedResponse(req2, rsp)
}

func (c *headerAuthClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

func (c *headerAuthClient) isUnauthorizedResponse(rsp *http.Response) bool {
	if c.isUnauthorized != nil {
		return c.isUnauthorized(rsp)
	}

	return rsp.StatusCode == http.StatusUnauthorized
}

// handleUnauthorizedResponse handles 401 responses or custom unauthorized conditions.
func (c *headerAuthClient) handleUnauthorizedResponse(
	req *http.Request,
	rsp *http.Response,
) (*http.Response, error) {
	if c.isUnauthorizedResponse(rsp) {
		if c.unauthorized != nil {
			return c.unauthorized(c.headers, req, rsp)
		}
	}

	return rsp, nil
}

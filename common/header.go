package common

import (
	"context"
	"net/http"
)

type HeaderAuthClientOption func(params *headerClientParams)

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

// WithDynamicHeaders sets a function that will be called on every request to
// get additional headers to use. Use this for things like time-based tokens
// or loading headers from some external authority. The function can access a
// copy of the request object to use its metadata for generating headers.
func WithDynamicHeaders(f DynamicHeadersGenerator) HeaderAuthClientOption {
	return func(params *headerClientParams) {
		params.dynamicHeaders = f
	}
}

type DynamicHeadersGenerator func(http.Request) ([]Header, error)

// oauthClientParams is the internal configuration for the oauth http client.
type headerClientParams struct {
	client         *http.Client
	headers        []Header
	dynamicHeaders DynamicHeadersGenerator
	debug          func(req *http.Request, rsp *http.Response)
	unauthorized   func(hdrs []Header, req *http.Request, rsp *http.Response) (*http.Response, error)
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
	}
}

type headerAuthClient struct {
	client         *http.Client
	headers        []Header
	dynamicHeaders DynamicHeadersGenerator
	debug          func(req *http.Request, rsp *http.Response)
	unauthorized   func(hdrs []Header, req *http.Request, rsp *http.Response) (*http.Response, error)
}

func (c *headerAuthClient) Do(req *http.Request) (*http.Response, error) {
	// This allows us to attach headers without modifying the input
	req = req.Clone(req.Context())

	for _, header := range c.headers {
		req.Header.Add(header.Key, header.Value)
	}

	if c.dynamicHeaders != nil {
		hdrs, err := c.dynamicHeaders(*req)
		if err != nil {
			return nil, err
		}

		for _, header := range hdrs {
			req.Header.Add(header.Key, header.Value)
		}
	}

	rsp, err := c.client.Do(req)
	if err != nil {
		return rsp, err
	}

	if c.debug != nil {
		c.debug(req, cloneResponse(rsp))
	}

	// Certain providers return 401 when the credentials has been invalidated.
	// This may indicate that the credentials needs to be forcefully refreshed.
	// Since this is per-provider, the caller can provide a custom handler.
	if rsp.StatusCode == http.StatusUnauthorized {
		if c.unauthorized != nil {
			return c.unauthorized(c.headers, req, rsp)
		}
	}

	return rsp, nil
}

func (c *headerAuthClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

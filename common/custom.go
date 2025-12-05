// nolint:revive
package common

import (
	"context"
	"net/http"
)

type CustomAuthClientOption func(params *customClientParams)

// NewCustomAuthHTTPClient returns a new http client, which will
// do generic header or query-param-based authentication. It does this by
// automatically adding the provided headers and/or query params to every
// request. There's no additional logic for refreshing tokens or anything like
// that. This is appropriate for APIs that require static but somewhat odd
// authentication mechanisms involving headers or query params.
func NewCustomAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	opts ...CustomAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	params := &customClientParams{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(params)
	}

	return newCustomAuthClient(ctx, params.prepare()), nil
}

// WithCustomClient sets the http client to use for the connector. Its usage is optional.
func WithCustomClient(client *http.Client) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.client = client
	}
}

// WithCustomDebug sets a debug function to be called on every request and response,
// after the response has been received from the downstream API.
func WithCustomDebug(f func(req *http.Request, rsp *http.Response)) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.debug = f
	}
}

// WithCustomHeaders sets the headers to use for the connector. Its usage is optional.
func WithCustomHeaders(headers ...Header) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.headers = append(params.headers, headers...)
	}
}

// WithCustomQueryParams sets the query params to use for the connector. Its usage is optional.
func WithCustomQueryParams(ps ...QueryParam) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.params = append(params.params, ps...)
	}
}

// WithCustomUnauthorizedHandler sets the function to call whenever the response is 401 unauthorized.
// This is useful for handling the case where the server has invalidated the credentials, and the client
// needs to refresh. It's optional.
func WithCustomUnauthorizedHandler(
	f func(hdrs []Header, params []QueryParam, req *http.Request, rsp *http.Response) (*http.Response, error),
) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.unauthorized = f
	}
}

// WithCustomIsUnauthorizedHandler sets the function to call
// whenever the response is unauthorized (not necessarily 401).
// This is useful for handling the case where the server has invalidated the token, and the client
// needs to forcefully refresh. It's optional.
func WithCustomIsUnauthorizedHandler(
	f func(rsp *http.Response) bool,
) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.isUnauthorized = f
	}
}

// WithCustomDynamicHeaders sets a function that will be called on every request to
// get additional headers to use. Use this for things like time-based tokens
// or loading headers from some external authority. The function can access a
// copy of the request object to use its metadata for generating headers.
func WithCustomDynamicHeaders(f DynamicHeadersGenerator) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.dynamicHeaders = f
	}
}

// WithCustomDynamicQueryParams sets a function that will be called on every request to
// get additional query params to use. Use this for things like time-based tokens
// or loading query params from some external authority. The function can access a
// copy of the request object to use its metadata for generating query params.
func WithCustomDynamicQueryParams(f DynamicQueryParamsGenerator) CustomAuthClientOption {
	return func(params *customClientParams) {
		params.dynamicParams = f
	}
}

// oauthClientParams is the internal configuration for the oauth http client.
type customClientParams struct {
	client         *http.Client
	headers        Headers
	params         QueryParams
	dynamicHeaders DynamicHeadersGenerator
	dynamicParams  DynamicQueryParamsGenerator
	debug          func(req *http.Request, rsp *http.Response)
	unauthorized   func(hdrs []Header, params []QueryParam, req *http.Request, rsp *http.Response) (*http.Response, error)
	isUnauthorized func(rsp *http.Response) bool
}

func (p *customClientParams) prepare() *customClientParams {
	if p.client == nil {
		p.client = http.DefaultClient
	}

	return p
}

// newCustomAuthClient returns a new http client for the connector, with automatic OAuth authentication.
func newCustomAuthClient(_ context.Context, params *customClientParams) AuthenticatedHTTPClient { //nolint:ireturn
	return &customAuthClient{
		client:         params.client,
		headers:        params.headers,
		dynamicHeaders: params.dynamicHeaders,
		params:         params.params,
		dynamicParams:  params.dynamicParams,
		debug:          params.debug,
		unauthorized:   params.unauthorized,
		isUnauthorized: params.isUnauthorized,
	}
}

type customAuthClient struct {
	client         *http.Client
	headers        Headers
	params         QueryParams
	dynamicHeaders DynamicHeadersGenerator
	dynamicParams  DynamicQueryParamsGenerator
	debug          func(req *http.Request, rsp *http.Response)
	unauthorized   func(hdrs []Header, params []QueryParam, req *http.Request, rsp *http.Response) (*http.Response, error)
	isUnauthorized func(rsp *http.Response) bool
}

func (c *customAuthClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

func (c *customAuthClient) getHeaders(req *http.Request) (Headers, error) { // nolint:funcorder
	var hdrs Headers

	if c.headers != nil {
		hdrs = append(hdrs, c.headers...)
	}

	if c.dynamicHeaders != nil {
		dhdrs, err := c.dynamicHeaders(req)
		if err != nil {
			return nil, err
		}

		hdrs = append(hdrs, dhdrs...)
	}

	return hdrs, nil
}

func (c *customAuthClient) getQueryParams(req *http.Request) (QueryParams, error) { // nolint:funcorder
	var params QueryParams

	if c.params != nil {
		params = append(params, c.params...)
	}

	if c.dynamicParams != nil {
		dparams, err := c.dynamicParams(req)
		if err != nil {
			return nil, err
		}

		params = append(params, dparams...)
	}

	return params, nil
}

func (c *customAuthClient) Do(req *http.Request) (*http.Response, error) {
	// This allows us to attach headers without modifying the input
	req2 := req.Clone(req.Context())

	hdrs, err := c.getHeaders(req2)
	if err != nil {
		return nil, err
	}

	if hdrs != nil {
		hdrs.ApplyToRequest(req2)
	}

	params, err := c.getQueryParams(req2)
	if err != nil {
		return nil, err
	}

	if params != nil {
		params.ApplyToRequest(req2)
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

func (c *customAuthClient) isUnauthorizedResponse(rsp *http.Response) bool {
	if c.isUnauthorized != nil {
		return c.isUnauthorized(rsp)
	}

	return rsp.StatusCode == http.StatusUnauthorized
}

// handleUnauthorizedResponse handles 401 responses or custom unauthorized conditions.
func (c *customAuthClient) handleUnauthorizedResponse(
	req *http.Request,
	rsp *http.Response,
) (*http.Response, error) {
	if c.isUnauthorizedResponse(rsp) {
		if c.unauthorized != nil {
			return c.unauthorized(c.headers, c.params, req, rsp)
		}
	}

	return rsp, nil
}

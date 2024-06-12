package common

import (
	"context"
	"net/http"
)

type QueryParam struct {
	Key   string
	Value string
}

type QueryParamAuthClientOption func(params *queryParamClientParams)

// NewQueryParamAuthHTTPClient returns a new http client, which will
// do generic query-param-based authentication. It does this by automatically
// adding the provided query params to every request. There's no additional
// logic for refreshing tokens or anything like that. This is appropriate
// for APIs that use keys encoded in the query params.
func NewQueryParamAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	opts ...QueryParamAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	params := &queryParamClientParams{}
	for _, opt := range opts {
		opt(params)
	}

	return newQueryParamAuthClient(ctx, params.prepare()), nil
}

// WithQueryParamClient sets the http client to use for the connector. Its usage is optional.
func WithQueryParamClient(client *http.Client) QueryParamAuthClientOption {
	return func(params *queryParamClientParams) {
		params.client = client
	}
}

func WithQueryParams(ps ...QueryParam) QueryParamAuthClientOption {
	return func(params *queryParamClientParams) {
		params.params = append(params.params, ps...)
	}
}

// WithQueryParamDebug sets a debug function to be called on every request and response,
// after the response has been received from the downstream API.
func WithQueryParamDebug(f func(req *http.Request, rsp *http.Response)) QueryParamAuthClientOption {
	return func(params *queryParamClientParams) {
		params.debug = f
	}
}

// queryParamClientParams is the internal configuration for the oauth http client.
type queryParamClientParams struct {
	client *http.Client
	params []QueryParam
	debug  func(req *http.Request, rsp *http.Response)
}

func (p *queryParamClientParams) prepare() *queryParamClientParams {
	if p.client == nil {
		p.client = http.DefaultClient
	}

	return p
}

// newQueryParamAuthClient returns a new http client for the connector, with automatic OAuth authentication.
func newQueryParamAuthClient(_ context.Context, params *queryParamClientParams) AuthenticatedHTTPClient { //nolint:ireturn
	return &queryParamAuthClient{
		client: params.client,
		params: params.params,
		debug:  params.debug,
	}
}

type queryParamAuthClient struct {
	client *http.Client
	params []QueryParam
	debug  func(req *http.Request, rsp *http.Response)
}

func (c *queryParamAuthClient) Do(req *http.Request) (*http.Response, error) {
	// This allows us to modify query params without mutating the input
	req = req.Clone(req.Context())

	query := req.URL.Query()
	for _, p := range c.params {
		query.Add(p.Key, p.Value)
	}

	req.URL.RawQuery = query.Encode()

	rsp, err := c.client.Do(req)
	if err != nil {
		return rsp, err
	}

	if c.debug != nil {
		c.debug(req, cloneResponse(rsp))
	}

	return rsp, nil
}

func (c *queryParamAuthClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

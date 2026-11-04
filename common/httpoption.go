// nolint:revive,godoclint
package common

import (
	"net/http"
)

// HTTPOptionIn identifies where an HTTPOption is attached on the request.
type HTTPOptionIn string

const (
	// HTTPOptionInHeader attaches the option as an HTTP request header.
	HTTPOptionInHeader HTTPOptionIn = "header"

	// HTTPOptionInQuery attaches the option as a URL query parameter.
	// Note that query parameter values end up in URLs, which are commonly
	// logged; prefer HTTPOptionInHeader for sensitive values.
	HTTPOptionInQuery HTTPOptionIn = "query"
)

// HTTPOption is a static key/value pair attached to every request made by an
// AuthenticatedHTTPClient, either as a header or a URL query parameter. It is
// meant for provider requirements beyond the primary auth credential, such as
// an extra client id or subscription key header, typically sourced from
// connection metadata.
type HTTPOption struct {
	In    HTTPOptionIn `json:"in"`
	Key   string       `json:"key"`
	Value string       `json:"value"`
}

// NewHTTPOptionsClient wraps an AuthenticatedHTTPClient so that the given
// options are attached to every request it makes. Options are applied
// set-if-missing: a value already present on the request (set by the caller
// or by a connector) wins, so nothing is overwritten or double-applied.
// If opts is empty, the client is returned unwrapped.
func NewHTTPOptionsClient(client AuthenticatedHTTPClient, opts []HTTPOption) AuthenticatedHTTPClient { //nolint:ireturn,lll
	if len(opts) == 0 {
		return client
	}

	var (
		headers     Headers
		queryParams QueryParams
	)

	for _, opt := range opts {
		switch opt.In {
		case HTTPOptionInHeader:
			headers = append(headers, Header{
				Key:   opt.Key,
				Value: opt.Value,
				Mode:  HeaderModeSetIfMissing,
			})
		case HTTPOptionInQuery:
			queryParams = append(queryParams, QueryParam{
				Key:   opt.Key,
				Value: opt.Value,
				Mode:  QueryParamModeSetIfMissing,
			})
		}
	}

	return &httpOptionsClient{
		client:      client,
		headers:     headers,
		queryParams: queryParams,
	}
}

// httpOptionsClient decorates an AuthenticatedHTTPClient with static headers
// and query parameters that are attached to every request.
type httpOptionsClient struct {
	client      AuthenticatedHTTPClient
	headers     Headers
	queryParams QueryParams
}

func (c *httpOptionsClient) Do(req *http.Request) (*http.Response, error) {
	// This allows us to attach headers/query params without mutating the input.
	req2 := req.Clone(req.Context())

	c.headers.ApplyToRequest(req2)
	c.queryParams.ApplyToRequest(req2)

	return c.client.Do(req2)
}

func (c *httpOptionsClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

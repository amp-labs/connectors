package common

import (
	"context"
)

// NewApiKeyHeaderAuthHTTPClient returns a new http client, with automatic API key
// authentication. Specifically this means that the client will automatically
// add the API key (as a header) to every request.
func NewApiKeyHeaderAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	headerName, apiKey string,
	opts ...HeaderAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	return NewHeaderAuthHTTPClient(ctx, append(opts, WithHeaders(Header{
		Key:   headerName,
		Value: apiKey,
	}))...)
}

// NewApiKeyQueryParamAuthHTTPClient returns a new http client, with automatic API key
// authentication. Specifically this means that the client will automatically
// add the API key (as a query param) to every request.
func NewApiKeyQueryParamAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	queryParamName, apiKey string,
	opts ...QueryParamAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	return NewQueryParamAuthHTTPClient(ctx, append(opts, WithQueryParams(QueryParam{
		Key:   queryParamName,
		Value: apiKey,
	}))...)
}

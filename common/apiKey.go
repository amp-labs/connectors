package common

import (
	"context"
)

// NewApiKeyHeaderAuthHTTPClient returns a new http client, with automatic API key
// authentication. Specifically this means that the client will automatically
// add the API key (as a header) to every request.
// HeaderValue must be in the correct format. This sometimes means adding prefix to the API Key.
func NewApiKeyHeaderAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	headerName, headerValue string,
	opts ...HeaderAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	return NewHeaderAuthHTTPClient(ctx, append(opts, WithHeaders(Header{
		Key:   headerName,
		Value: headerValue,
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

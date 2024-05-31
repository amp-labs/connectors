package common

import (
	"context"
)

// NewApiKeyAuthHTTPClient returns a new http client, with automatic API key
// authentication. Specifically this means that the client will automatically
// add the API key (as a header) to every request.
func NewApiKeyAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	headerName, apiKey string,
	opts ...HeaderAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	return NewHeaderAuthHTTPClient(ctx, append(opts, WithHeaders(Header{
		Key:   headerName,
		Value: apiKey,
	}))...)
}

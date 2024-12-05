package common

import (
	"context"
)

// NewJwtAuthHTTPClient returns a new http client, with automatic Jwt generation &
// authentication. Specifically this means that the client will automatically
// add the Jwt (as a header) to every request, based on the given secret and claims.
func NewJwtAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	opts ...HeaderAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	return NewHeaderAuthHTTPClient(ctx, opts...)
}

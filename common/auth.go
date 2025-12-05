// nolint:revive
package common

import (
	"context"
	"errors"
	"net/http"
)

// authTokenContextKey is the type used for the context key for the auth token.
type authTokenContextKey string

// authTokenKey is the context key for the auth token.
const authTokenKey = authTokenContextKey("authToken")

// AuthenticatedHTTPClient is an interface for an http client which can automatically
// authenticate itself. This is useful for OAuth authentication, where the access token
// needs to be refreshed automatically. The signatures are a subset of http.Client,
// so it can be used as a (mostly) drop-in replacement.
type AuthenticatedHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

// AuthToken is a type alias for a string representing an authentication token.
// This can be used to store the token in a context.
type AuthToken string

func (t AuthToken) String() string {
	return string(t)
}

// ErrMissingAccessToken returned when access token was not present in the context.Context.
var ErrMissingAccessToken = errors.New("missing access token")

// WithAuthToken returns a new context with the given auth token.
func WithAuthToken(ctx context.Context, token AuthToken) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, authTokenKey, token)
}

// GetAuthToken returns the auth token from the context, if it exists.
func GetAuthToken(ctx context.Context) (AuthToken, bool) {
	if ctx == nil {
		return "", false
	}

	sub := ctx.Value(authTokenKey)
	if sub == nil {
		return "", false
	}

	token, ok := sub.(AuthToken)
	if !ok {
		return "", false
	}

	return token, true
}

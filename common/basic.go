// nolint:revive,godoclint
package common

import (
	"context"
	"encoding/base64"
)

// NewBasicAuthHTTPClient returns a new http client, with automatic Basic authentication.
// Specifically this means that the client will automatically add the Basic auth header
// to every request. The username and password are provided as arguments.
func NewBasicAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	user, pass string,
	opts ...HeaderAuthClientOption,
) (AuthenticatedHTTPClient, error) {
	return NewHeaderAuthHTTPClient(ctx, append(opts, WithHeaders(Header{
		Key:   "Authorization",
		Value: "Basic " + basicAuth(user, pass),
	}))...)
}

// shamelessly stolen from https://pkg.go.dev/net/http#Request.SetBasicAuth
func basicAuth(username, password string) string {
	auth := username + ":" + password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}

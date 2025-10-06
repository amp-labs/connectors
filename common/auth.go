package common

import "net/http"

// AuthenticatedHTTPClient is an interface for an http client which can automatically
// authenticate itself. This is useful for OAuth authentication, where the access token
// needs to be refreshed automatically. The signatures are a subset of http.Client,
// so it can be used as a (mostly) drop-in replacement.
type AuthenticatedHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

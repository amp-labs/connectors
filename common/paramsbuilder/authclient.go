package paramsbuilder

import (
	"errors"

	"github.com/amp-labs/connectors/common"
)

var ErrMissingClient = errors.New("http client not set")

// ClientHolder gives authenticated client.
// This is useful to check if object is of interface type to get access to HTTP client.
type ClientHolder interface {
	GiveClient() *AuthClient
}

// AuthClient params sets up authenticated proxy HTTP client.
type AuthClient struct {
	// Caller is an HTTP client that knows how to make authenticated requests.
	// It also knows how to handle authentication and API response errors.
	Caller *common.HTTPClient
}

func (p *AuthClient) ValidateParams() error {
	// http client must be defined.
	if p.Caller == nil {
		return ErrMissingClient
	}

	// authentication client should be present.
	if p.Caller.Client == nil {
		return ErrMissingClient
	}

	return nil
}

// WithAuthenticatedClient sets up an HTTP client that uses your implementation of authentication.
func (p *AuthClient) WithAuthenticatedClient(client common.AuthenticatedHTTPClient) {
	p.Caller = &common.HTTPClient{
		Client:       client,
		ErrorHandler: common.InterpretError,
	}
}

func (p *AuthClient) GiveClient() *AuthClient {
	return p
}

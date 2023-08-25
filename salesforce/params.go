package salesforce

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the salesforce connector configuration.
type Option func(params *sfParams)

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token) Option {
	return func(params *sfParams) {
		oauthClient, err := common.NewOAuthHTTPClient(ctx,
			common.WithClient(client),
			common.WithOAuthConfig(config),
			common.WithOAuthToken(token))
		if err != nil {
			panic(err) // caught in NewConnector
		}

		WithAuthenticatedClient(oauthClient)(params)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *sfParams) {
		params.client = &common.JSONHTTPClient{
			Client:       client,
			ErrorHandler: common.InterpretError,
		}
	}
}

// WithSubdomain sets the salesforce subdomain to use for the connector. It's required.
func WithSubdomain(workspaceRef string) Option {
	return func(params *sfParams) {
		params.subdomain = workspaceRef
	}
}

// sfParams is the internal configuration for the salesforce connector.
type sfParams struct {
	client    *common.JSONHTTPClient // required
	subdomain string                 // required
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *sfParams) prepare() (out *sfParams, err error) {
	if p.client == nil {
		return nil, ErrMissingClient
	}

	if len(p.subdomain) == 0 {
		return nil, ErrMissingSubdomain
	}

	return p, nil
}

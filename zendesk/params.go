package zendesk

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the zendesk connector configuration.
type Option func(params *zendeskParams)

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
	opts ...common.OAuthOption,
) Option {
	return func(params *zendeskParams) {
		options := []common.OAuthOption{
			common.WithClient(client),
			common.WithOAuthConfig(config),
			common.WithOAuthToken(token),
		}

		oauthClient, err := common.NewOAuthHTTPClient(ctx, append(options, opts...)...)
		if err != nil {
			panic(err) // caught in NewConnector
		}

		WithAuthenticatedClient(oauthClient)(params)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *zendeskParams) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:       client,
				ErrorHandler: common.InterpretError,
			},
		}
	}
}

// WithWorkspace sets the zendesk workspace to use for the connector. It's required.
func WithWorkspace(workspaceRef string) Option {
	return func(params *zendeskParams) {
		params.workspace = workspaceRef
	}
}

// zendeskParams is the internal configuration for the zendesk connector.
type zendeskParams struct {
	client    *common.JSONHTTPClient // required
	workspace string                 // required
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *zendeskParams) prepare() (out *zendeskParams, err error) {
	if p.client == nil {
		return nil, ErrMissingClient
	}

	if len(p.workspace) == 0 {
		return nil, ErrMissingWorkspace
	}

	return p, nil
}

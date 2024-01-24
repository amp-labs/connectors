package hubspot

import (
	"context"
	"net/http"
	"path"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

const (
	// DefaultPageSize is the default page size for paginated requests.
	DefaultPageSize = "100"
)

// Option is a function which mutates the hubspot connector configuration.
type Option func(params *hubspotParams)

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
	opts ...common.OAuthOption,
) Option {
	return func(params *hubspotParams) {
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
	return func(params *hubspotParams) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:       client,
				ErrorHandler: common.InterpretError,
			},
		}
	}
}

// WithModule sets the hubspot API module to use for the connector. It's required.
func WithModule(module APIModule) Option {
	return func(params *hubspotParams) {
		elems := []string{module.Label, module.Version}
		params.module = path.Join(elems...)
	}
}

// hubspotParams is the internal configuration for the hubspot connector.
type hubspotParams struct {
	client *common.JSONHTTPClient // required
	module string                 // required
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *hubspotParams) prepare() (out *hubspotParams, err error) {
	if p.client == nil {
		return nil, ErrMissingClient
	}

	if len(p.module) == 0 {
		return nil, ErrMissingAPIModule
	}

	return p, nil
}

func requiresFiltering(config common.ReadParams) bool {
	return !config.Since.IsZero()
}

package connector

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

var (
	// ErrMissingClient is returned when a connector is created without a client.
	ErrMissingClient = errors.New("missing client")

	// ErrMissingProvider is returned when a connector is created without a provider.
	ErrMissingProvider = errors.New("missing provider")
)

type Option func(*connectorParams)

type connectorParams struct {
	provider      providers.Provider
	client        *common.JSONHTTPClient
	substitutions map[string]string
}

func (p *connectorParams) prepare() (*connectorParams, error) {
	if p.provider == "" {
		return nil, ErrMissingProvider
	}

	if p.client == nil {
		return nil, ErrMissingClient
	}

	return p, nil
}

// WithCatalogSubstitutions sets the provider values to use while making substitutions &
// reading from providers.yaml. If the provider values are not set, the connector
// will use error out.
func WithCatalogSubstitutions(substitutions map[string]string) Option {
	return func(params *connectorParams) {
		params.substitutions = substitutions
	}
}

// WithClient sets the http client to use for the connector.
func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
	opts ...common.OAuthOption,
) Option {
	return func(params *connectorParams) {
		options := []common.OAuthOption{
			common.WithOAuthClient(client),
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
	return func(params *connectorParams) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:       client,
				ErrorHandler: common.InterpretError,
			},
		}
	}
}

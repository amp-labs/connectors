package proxy

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

// ErrMissingClient is returned when a connector is created without a client.
var ErrMissingClient = errors.New("missing client")

// WithCatalogSubstitutions sets the provider values to use while making substitutions &
// reading from providers.yaml. If the provider values are not set, the connector
// will use error out.
func WithCatalogSubstitutions(substitutions map[string]string) Option {
	return func(conn *Connector) {
		conn.substitutions = &substitutions
	}
}

// WithClient sets the http client to use for the connector.
func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
	opts ...common.OAuthOption,
) Option {
	return func(conn *Connector) {
		options := []common.OAuthOption{
			common.WithClient(client),
			common.WithOAuthConfig(config),
			common.WithOAuthToken(token),
		}

		oauthClient, err := common.NewOAuthHTTPClient(ctx, append(options, opts...)...)
		if err != nil {
			panic(err) // caught in NewConnector
		}

		WithAuthenticatedClient(oauthClient)(conn)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(conn *Connector) {
		conn.Client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:       client,
				ErrorHandler: common.InterpretError,
			},
		}
	}
}

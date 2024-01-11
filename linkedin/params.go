package linkedin

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

var ErrMissingClient = errors.New("missing http client")

// Option is a function which mutates the hubspot connector configuration.
type Option func(params *linkedInParams)

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
	opts ...common.OAuthOption,
) Option {
	return func(params *linkedInParams) {
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
	return func(params *linkedInParams) {
		params.client = &common.HTTPClient{
			Client:       client,
			ErrorHandler: common.InterpretError,
		}
	}
}

// linkedInParams is the internal configuration for the hubspot connector.
type linkedInParams struct {
	client *common.HTTPClient // required
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *linkedInParams) prepare() (out *linkedInParams, err error) {
	if p.client == nil {
		return nil, ErrMissingClient
	}

	return p, nil
}

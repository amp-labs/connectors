package common

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var ClientCredentials Oauth2OptsGrantType = "clientcredentials" //nolint:gochecknoglobals

// get2leggedTokenSource generates tokenSource using the clientcreds config.
func get2leggedTokenSource(ctx context.Context, params *oauthClientParams) oauth2.TokenSource { //nolint:ireturn
	if params.tokenSource != nil {
		return params.tokenSource
	}

	return params.clientCredsConfig.TokenSource(ctx)
}

// WithClientCredentialsConfig sets the client-credentials config to use for the connector.
// It's required in two-legged OAuth2, unless TokenSource is provided.
func WithClientCredentialsConfig(config *clientcredentials.Config) OAuthOption {
	return func(params *oauthClientParams) {
		params.clientCredsConfig = config
	}
}

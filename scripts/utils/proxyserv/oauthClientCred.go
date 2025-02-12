// nolint:ireturn
package proxyserv

import (
	"context"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2/clientcredentials"
)

func (f Factory) CreateProxyOAuth2ClientCreds(ctx context.Context) *Proxy {
	params := createClientAuthParams(f.Provider, f.Registry)
	providerInfo := getProviderConfig(f.Provider, f.CatalogVariables)
	cfg := configureOAuthClientCredentials(params.ID, params.Secret, params.Scopes, providerInfo)
	httpClient := setupOAuth2ClientCredentialsHTTPClient(ctx, providerInfo, cfg, f.Debug)
	baseURL := getBaseURL(providerInfo)

	return newProxy(baseURL, httpClient)
}

func configureOAuthClientCredentials(
	clientId, clientSecret string, scopes []string, providerInfo *providers.ProviderInfo,
) *clientcredentials.Config {
	cfg := &clientcredentials.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		TokenURL:     providerInfo.Oauth2Opts.TokenURL,
	}

	if providerInfo.Oauth2Opts.ExplicitScopesRequired {
		cfg.Scopes = scopes
	}

	if providerInfo.Oauth2Opts.Audience != nil {
		aud := providerInfo.Oauth2Opts.Audience
		cfg.EndpointParams = url.Values{"audience": aud}
	}

	return cfg
}

func setupOAuth2ClientCredentialsHTTPClient(
	ctx context.Context, prov *providers.ProviderInfo, cfg *clientcredentials.Config, debug bool,
) common.AuthenticatedHTTPClient {
	client, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: debug,
		OAuth2ClientCreds: &providers.OAuth2ClientCredentialsParams{
			Config: cfg,
		},
	})
	if err != nil {
		panic(err)
	}

	cc, err := connector.NewConnector(prov.Name, connector.WithAuthenticatedClient(client))
	if err != nil {
		panic(err)
	}

	return cc.HTTPClient().Client
}

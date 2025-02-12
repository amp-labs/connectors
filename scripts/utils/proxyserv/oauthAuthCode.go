// nolint:ireturn
package proxyserv

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

func (f Factory) CreateProxyOAuth2AuthCode(ctx context.Context) *Proxy {
	params := createClientAuthParams(f.Provider, f.Registry)
	tokens := getTokensFromRegistry(f.CredsFilePath)
	providerInfo := getProviderConfig(f.Provider, f.CatalogVariables)
	cfg := configureOAuthAuthCode(params.ID, params.Secret, params.Scopes, providerInfo)
	httpClient := setupOAuth2AuthCodeHTTPClient(ctx, providerInfo, cfg, tokens, f.Debug)
	baseURL := getBaseURL(providerInfo)

	return newProxy(baseURL, httpClient)
}

func configureOAuthAuthCode(
	clientId, clientSecret string, scopes []string, providerInfo *providers.ProviderInfo,
) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   providerInfo.Oauth2Opts.AuthURL,
			TokenURL:  providerInfo.Oauth2Opts.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}

// This helps with refreshing tokens automatically.
func setupOAuth2AuthCodeHTTPClient(
	ctx context.Context, prov *providers.ProviderInfo, cfg *oauth2.Config, tokens *oauth2.Token, debug bool,
) common.AuthenticatedHTTPClient {
	client, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: debug,
		OAuth2AuthCodeCreds: &providers.OAuth2AuthCodeParams{
			Config: cfg,
			Token:  tokens,
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

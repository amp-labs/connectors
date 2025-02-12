// nolint:ireturn
package proxyserv

import (
	"context"
	"fmt"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/providers"
)

func (f Factory) CreateProxyAPIKey(ctx context.Context) *Proxy {
	apiKey := getAPIKey(f.Registry)
	providerInfo := getProviderConfig(f.Provider, f.CatalogVariables)
	httpClient := setupAPIKeyHTTPClient(ctx, providerInfo, apiKey, f.Debug)
	baseURL := getBaseURL(providerInfo)

	return newProxy(baseURL, httpClient)
}

func setupAPIKeyHTTPClient(
	ctx context.Context, prov *providers.ProviderInfo, apiKey string, debug bool,
) common.AuthenticatedHTTPClient {
	client, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug:  debug,
		ApiKey: apiKey,
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

func getAPIKey(registry scanning.Registry) string {
	apiKey := registry.MustString(credscanning.Fields.ApiKey.Name)
	if apiKey == "" {
		_, _ = fmt.Fprintln(os.Stderr, "api key from registry is empty")
		os.Exit(1)
	}

	return apiKey
}

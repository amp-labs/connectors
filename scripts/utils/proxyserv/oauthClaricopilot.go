// nolint:ireturn
package proxyserv

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/generic"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/claricopilot"
)

func (f Factory) CreateProxyClariCopilotAPIKey(ctx context.Context) *Proxy {
	providerInfo := getProviderConfig(f.Provider, f.CatalogVariables)
	httpClient := setupClariCopilotAPIKeyHTTPClient(ctx, providerInfo, f.Substitutions)
	baseURL := getBaseURL(providerInfo)

	return newProxy(baseURL, httpClient)
}

func setupClariCopilotAPIKeyHTTPClient(
	ctx context.Context, prov *providers.ProviderInfo,
	metadata map[string]string,
) common.AuthenticatedHTTPClient {
	client, err := claricopilot.NewClariCopilotAuthHTTPClient(ctx, "X-Api-Key", "8nfpwD7uJe4RwMS02XwqN3kcQScHQnt34Trd9lfu", "X-Api-Password", "2087bc28-1158-4187-afee-90f207081e61")
	if err != nil {
		panic(err)
	}

	cc, err := generic.NewConnector(prov.Name,
		generic.WithAuthenticatedClient(client),
		generic.WithMetadata(metadata),
	)
	if err != nil {
		panic(err)
	}

	return cc.HTTPClient().Client
}

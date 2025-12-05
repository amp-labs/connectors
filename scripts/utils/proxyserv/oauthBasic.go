package proxyserv

import (
	"context"
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/generic"
	"github.com/amp-labs/connectors/providers"
)

func (f Factory) CreateProxyBasic(ctx context.Context) *Proxy {
	params := createBasicParams(f.Registry)
	providerInfo := getProviderConfig(f.Provider, f.CatalogVariables)
	httpClient := setupBasicAuthHTTPClient(ctx, providerInfo, params.User, params.Pass, f.Debug, f.Metadata)
	baseURL := getBaseURL(providerInfo)

	return newProxy(baseURL, httpClient)
}

func createBasicParams(registry scanning.Registry) *providers.BasicParams {
	user := registry.MustString(credscanning.Fields.Username.Name)
	pass := registry.MustString(credscanning.Fields.Password.Name)

	if len(user)+len(pass) == 0 {
		log.Fatalf("Missing username or password")
	}

	if len(user) == 0 {
		slog.Warn("no username for basic authentication, ensure that it is not required")
	}

	if len(pass) == 0 {
		slog.Warn("no password for basic authentication, ensure that it is not required")
	}

	return &providers.BasicParams{
		User: user,
		Pass: pass,
	}
}

func setupBasicAuthHTTPClient(
	ctx context.Context, prov *providers.ProviderInfo, user, pass string, debug bool,
	metadata map[string]string,
) common.AuthenticatedHTTPClient {
	client, err := prov.NewClient(ctx, &providers.NewClientParams{
		Debug: debug,
		BasicCreds: &providers.BasicParams{
			User: user,
			Pass: pass,
		},
	})
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

package netsuite

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/netsuite"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetNetsuiteConnector(ctx context.Context) *netsuite.Connector {
	return getNetsuiteConnector(ctx, providers.NetsuiteModuleRESTAPI)
}

func GetNetsuiteRESTAPIConnector(ctx context.Context) *netsuite.Connector {
	return getNetsuiteConnector(ctx, providers.NetsuiteModuleRESTAPI)
}

func GetNetsuiteSuiteQLConnector(ctx context.Context) *netsuite.Connector {
	return getNetsuiteConnector(ctx, providers.NetsuiteModuleSuiteQL)
}

func getNetsuiteConnector(ctx context.Context, module common.ModuleID) *netsuite.Connector {
	reader := getNetsuiteJSONReader()

	conn, err := netsuite.NewConnector(common.ConnectorParams{
		AuthenticatedClient: getAuthenticatedClient(ctx, reader),
		Workspace:           reader.Get(credscanning.Fields.Workspace),
		Module:              module,
	})
	if err != nil {
		utils.Fail("error creating netsuite connector", "error", err)
	}

	return conn
}

func getAuthenticatedClient(ctx context.Context, reader *credscanning.ProviderCredentials) common.AuthenticatedHTTPClient {
	options := []common.OAuthOption{
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(&oauth2.Token{
			AccessToken:  reader.Get(credscanning.Fields.AccessToken),
			RefreshToken: reader.Get(credscanning.Fields.RefreshToken),
		}),
	}

	client, err := common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		utils.Fail("error creating oauth client", "error", err)
	}

	return client
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	workspace := reader.Get(credscanning.Fields.Workspace)

	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://" + workspace + ".app.netsuite.com/app/login/oauth2/authorize.nl",
			TokenURL:  "https://" + workspace + ".suitetalk.api.netsuite.com/services/rest/auth/oauth2/v1/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"restlets",
			"rest_webservices",
		},
	}

	return cfg
}

func getNetsuiteJSONReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.Netsuite)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	return reader
}

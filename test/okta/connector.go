package okta

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/okta"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetOktaConnector(ctx context.Context) *okta.Connector {
	filePath := credscanning.LoadPath(providers.Okta)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	workspace := reader.Get(credscanning.Fields.Workspace)

	conn, err := okta.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig(workspace)),
		Workspace:           workspace,
	})
	if err != nil {
		utils.Fail("error creating okta connector", "error", err)
	}

	return conn
}

func getConfig(workspace string) func(*credscanning.ProviderCredentials) *oauth2.Config {
	return func(reader *credscanning.ProviderCredentials) *oauth2.Config {
		return &oauth2.Config{
			ClientID:     reader.Get(credscanning.Fields.ClientId),
			ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
			RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://" + workspace + ".okta.com/oauth2/v1/authorize",
				TokenURL:  "https://" + workspace + ".okta.com/oauth2/v1/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes: []string{
				"okta.users.read",
				"okta.groups.read",
				"okta.apps.read",
			},
		}
	}
}

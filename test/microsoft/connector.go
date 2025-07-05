package dynamicscrm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/microsoft"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetMicrosoftGraphConnector(ctx context.Context) *microsoft.Connector {
	filePath := credscanning.LoadPath(providers.Microsoft)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := microsoft.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
		},
	)
	if err != nil {
		utils.Fail("error creating microsoft CRM connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	workspace := reader.Get(credscanning.Fields.Workspace)

	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://login.microsoftonline.com/" + workspace + "/oauth2/v2.0/authorize",
			TokenURL:  "https://login.microsoftonline.com/" + workspace + "/oauth2/v2.0/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"User.Read",
			"offline_access",
		},
	}
}

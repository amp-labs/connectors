package webflow

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	connector "github.com/amp-labs/connectors/providers/webflow"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetWebflowConnector(ctx context.Context) *connector.Connector {
	filePath := credscanning.LoadPath(providers.Webflow)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := connector.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		},
	)
	if err != nil {
		utils.Fail("error creating webflow connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://webflow.com/oauth/authorize",
			TokenURL:  "https://api.webflow.com/oauth/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

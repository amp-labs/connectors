package bentley

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/bentley"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetBentleyConnector(ctx context.Context) *bentley.Connector {
	filePath := credscanning.LoadPath(providers.Bentley)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := bentley.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		},
	)
	if err != nil {
		utils.Fail("error creating Bentley connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://ims.bentley.com/connect/authorize",
			TokenURL:  "https://ims.bentley.com/connect/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

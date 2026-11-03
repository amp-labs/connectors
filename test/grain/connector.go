package grain

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/grain"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetGrainConnector(ctx context.Context) *grain.Connector {
	filePath := credscanning.LoadPath(providers.Grain)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := grain.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
	})
	if err != nil {
		utils.Fail("error creating Grain connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		// Grain rejects localhost callbacks for this client, so tokens are minted
		// through the production callback; the redirect is unused on refresh.
		RedirectURL: "https://api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://grain.com/_/public-api/oauth2/authorize",
			TokenURL:  "https://api.grain.com/_/public-api/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

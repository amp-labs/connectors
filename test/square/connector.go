package square

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/square"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetSquareConnector(ctx context.Context) *square.Connector {
	filePath := credscanning.LoadPath(providers.SquareSandbox)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := square.NewSandboxConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		},
	)
	if err != nil {
		utils.Fail("error creating Square connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/oauth2/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://connect.squareupsandbox.com/oauth2/authorize",
			TokenURL:  "https://connect.squareupsandbox.com/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

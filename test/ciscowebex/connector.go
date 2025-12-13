package ciscowebex

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/ciscowebex"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetCiscoWebexConnector(ctx context.Context) *ciscowebex.Connector {
	filePath := credscanning.LoadPath(providers.CiscoWebex)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := ciscowebex.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://webexapis.com/v1/authorize",
			TokenURL:  "https://webexapis.com/v1/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

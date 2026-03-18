package greenhouse

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/greenhouse"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetGreenhouseConnector(ctx context.Context) *greenhouse.Connector {
	filePath := credscanning.LoadPath(providers.Greenhouse)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := greenhouse.NewConnector(
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
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://auth.greenhouse.io/authorize",
			TokenURL:  "https://auth.greenhouse.io/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}

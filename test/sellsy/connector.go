package sellsy

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/sellsy"
	"github.com/amp-labs/connectors/test/utils"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetSellsyConnector(ctx context.Context) *sellsy.Connector {
	filePath := credscanning.LoadPath(providers.Sellsy)
	reader := testUtils.MustCreateProvCredJSON(filePath, true)

	conn, err := sellsy.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		},
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://login.sellsy.com/oauth2/authorization",
			TokenURL:  "https://login.sellsy.com/oauth2/access-tokens",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}

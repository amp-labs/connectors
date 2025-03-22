package monday

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/monday"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetMondayConnector(ctx context.Context) *monday.Connector {
	filePath := credscanning.LoadPath(providers.Monday)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", reader.Get(credscanning.Fields.ApiKey))

	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := monday.NewConnector(
		common.Parameters{AuthenticatedClient: client},
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
			TokenURL:  "https://auth.monday.com/oauth2/authorize",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}

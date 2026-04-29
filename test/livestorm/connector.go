package livestorm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/livestorm"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetLivestormConnector(ctx context.Context) *livestorm.Connector {
	filePath := credscanning.LoadPath(providers.Livestorm)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := livestorm.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
	})
	if err != nil {
		utils.Fail("error creating Livestorm connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.livestorm.co/oauth/authorize",
			TokenURL:  "https://app.livestorm.co/oauth/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
}

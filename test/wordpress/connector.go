package wordpress

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/wordpress"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetWordpressConnector(ctx context.Context) *wordpress.Connector {
	filePath := credscanning.LoadPath(providers.WordPress)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := wordpress.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
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
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL:  "https://public-api.wordpress.com/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

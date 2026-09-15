package bambooHR

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	bamboohr "github.com/amp-labs/connectors/providers/bambooHR"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetBambooHRConnector(ctx context.Context) *bamboohr.Connector {
	filePath := credscanning.LoadPath(providers.BambooHR)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := bamboohr.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
		},
	)
	if err != nil {
		utils.Fail("error creating bambooHR connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://{{.workspace}}.bamboohr.com/authorize.php?request=authorize",
			TokenURL:  "https://{{.workspace}}.bamboohr.com/token.php?request=token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

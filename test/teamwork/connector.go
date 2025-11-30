package teamwork

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/teamwork"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetTeamworkConnector(ctx context.Context) *teamwork.Connector {
	filePath := credscanning.LoadPath(providers.Teamwork)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := teamwork.NewConnector(
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
			AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:  "https://oauth2.googleapis.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{},
	}
}

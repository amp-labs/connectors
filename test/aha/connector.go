package aha

import (
	"context"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aha"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetAhaConnector(ctx context.Context) *aha.Connector {
	filePath := credscanning.LoadPath(providers.Aha)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := aha.NewConnector(
		parameters.Connector{
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

		RedirectURL: "https://api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://{{.workspace}}.aha.io/oauth/authorize",
			TokenURL:  "https://{{.workspace}}.aha.io/oauth/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}

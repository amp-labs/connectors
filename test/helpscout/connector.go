package helpscout

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/helpscout"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetHelpScoutConnector(ctx context.Context) *helpscout.Connector {
	filePath := credscanning.LoadPath(providers.HelpScoutMailbox)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := helpscout.NewConnector(
		helpscout.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
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
			TokenURL:  "https://api.helpscout.net/v2/oauth2/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}

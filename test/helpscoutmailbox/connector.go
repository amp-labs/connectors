package helpscoutmailbox

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/helpscoutmailbox"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetHelpScoutConnector(ctx context.Context) *helpscoutmailbox.Connector {
	filePath := credscanning.LoadPath(providers.HelpScoutMailbox)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := helpscoutmailbox.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
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

package snapchatads

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/snapchatads"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetConnector(ctx context.Context) *snapchatads.Connector {
	filePath := credscanning.LoadPath(providers.SnapchatAds)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)

	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	conn, err := snapchatads.NewConnector(common.ConnectorParams{
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
			AuthURL:   "https://accounts.snapchat.com/login/oauth2/authorize",
			TokenURL:  "https://accounts.snapchat.com/login/oauth2/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"snapchat-marketing-api",
		},
	}

	return &cfg
}

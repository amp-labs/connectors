package pinterest

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pinterest"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetConnector(ctx context.Context) *pinterest.Connector {
	filePath := credscanning.LoadPath(providers.Pinterest)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	conn, err := pinterest.NewConnector(common.ConnectorParams{
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
			AuthURL:   "https://www.pinterest.com/oauth",
			TokenURL:  "https://api.pinterest.com/v5/oauth/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		Scopes: []string{
			"ads:read",
			"ads:write",
			"boards:read",
			"boards:read_secret",
			"boards:write",
			"boards:write_secret",
			"pins:read",
			"pins:read_secret",
			"pins:write",
			"pins:write_secret",
			"user_accounts:read",
			"catalogs:read",
			"catalogs:write",
			"biz_access:read",
		},
	}

	return &cfg
}

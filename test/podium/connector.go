package podium

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/podium"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetConnector(ctx context.Context) *podium.Connector {
	filePath := credscanning.LoadPath(providers.Podium)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := podium.NewConnector(common.Parameters{
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
			AuthURL:   "https://api.podium.com/oauth/authorize",
			TokenURL:  "https://api.podium.com/oauth/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}

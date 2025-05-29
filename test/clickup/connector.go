package clickup

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/clickup"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetClickupConnector(ctx context.Context) *clickup.Connector {
	filePath := credscanning.LoadPath(providers.ClickUp)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	options := []common.OAuthOption{
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(&oauth2.Token{
			AccessToken:  reader.Get(credscanning.Fields.AccessToken),
			RefreshToken: reader.Get(credscanning.Fields.RefreshToken),
		}),
	}

	client, err := common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := clickup.NewConnector(
		parameters.Connector{
			AuthenticatedClient: client,
		},
	)
	if err != nil {
		utils.Fail("error creating clickup connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.clickup.com/api",
			TokenURL:  "https://api.clickup.com/api/v2/oauth/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		RedirectURL: "http://localhost:8080/callbacks/v1/oauth",
	}
}

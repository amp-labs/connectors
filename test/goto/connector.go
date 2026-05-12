package goto_test

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	gotoconn "github.com/amp-labs/connectors/providers/goto"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetGoToConnector(ctx context.Context, module common.ModuleID) *gotoconn.Connector {
	filePath := credscanning.LoadPath(providers.GoTo)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := gotoconn.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Module:              module,
		Metadata: map[string]string{
			"accountKey": "5276072959790856388",
		},
	})
	if err != nil {
		utils.Fail("create goto connector", "error: ", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://authentication.logmeininc.com/oauth/authorize",
			TokenURL:  "https://authentication.logmeininc.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

package xero

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/xero"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetXeroConnector(ctx context.Context) *xero.Connector {
	filePath := credscanning.LoadPath(providers.Xero)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)

	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := xero.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("create xero connector", "error: ", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {

	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://login.xero.com/identity/connect/authorize",
			TokenURL:  "https://identity.xero.com/connect/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

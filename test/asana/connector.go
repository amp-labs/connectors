package asana

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/asana"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetAsanaConnector(ctx context.Context) *asana.Connector {
	filePath := credscanning.LoadPath(providers.Asana)

	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := asana.NewConnector(
		asana.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating asana connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.asana.com/-/oauth_authorize",
			TokenURL:  "https://app.asana.com/-/oauth_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

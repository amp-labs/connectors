package kit

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/kit"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetKitConnector(ctx context.Context) *kit.Connector {
	filePath := credscanning.LoadPath(providers.Kit)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := kit.NewConnector(
		kit.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating kit connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.kit.com/authorize",
			TokenURL:  "https://app.kit.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

package gong

import (
	"context"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/gong"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
	"net/http"
)

func GetGongConnector(ctx context.Context) *gong.Connector {
	filePath := credscanning.LoadPath(providers.Gong)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := gong.NewConnector(
		gong.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating Gong connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.gong.io/oauth2/authorize",
			TokenURL:  "https://app.gong.io/oauth2/generate-customer-token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"api:calls:read:basic",
			"api:users:read",
			"api:calls:create:basic",
			"api:calls:read:basic",
			"api:meetings:user:delete",
			"api:meetings:user:update",
			"api:logs:read",
			"api:meetings:user:create",
			"api:workspaces:read",
		},
	}

	return cfg
}

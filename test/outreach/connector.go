package outreach

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/outreach"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetOutreachConnector(ctx context.Context) *outreach.Connector {
	filePath := credscanning.LoadPath(providers.Outreach)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := outreach.NewConnector(
		outreach.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating outreach connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://api.outreach.io/oauth/authorize",
			TokenURL:  "https://api.outreach.io/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"users.all",
			"accounts.read",
			"calls.all",
			"events.all",
			"teams.all",
		},
	}

	return cfg
}

package journeysapp

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/constantcontact"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetConstantContactConnector(ctx context.Context) *constantcontact.Connector {
	filePath := credscanning.LoadPath(providers.ConstantContact)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := constantcontact.NewConnector(
		constantcontact.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating Constant Contract connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://authz.constantcontact.com/oauth2/default/v1/authorize",
			TokenURL:  "https://authz.constantcontact.com/oauth2/default/v1/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: "http://localhost:8080/callbacks/v1/oauth",
		Scopes: []string{
			"account_read",
			"account_update",
			"contact_data",
			"offline_access",
			"campaign_data",
		},
	}
}

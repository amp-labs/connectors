package dynamicscrm

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/keap"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetKeapConnector(ctx context.Context) *keap.Connector {
	filePath := credscanning.LoadPath(providers.Keap)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := keap.NewConnector(
		keap.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating microsoft CRM connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.infusionsoft.com/app/oauth/authorize",
			TokenURL:  "https://api.infusionsoft.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

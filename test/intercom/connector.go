package intercom

import (
	"context"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
	"net/http"

	"github.com/amp-labs/connectors/providers/intercom"
	"github.com/amp-labs/connectors/test/utils"
)

func GetIntercomConnector(ctx context.Context) *intercom.Connector {
	filePath := credscanning.LoadPath(providers.Intercom)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := intercom.NewConnector(
		intercom.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating Intercom connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.intercom.com/oauth",
			TokenURL:  "https://api.intercom.io/auth/eagle/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{},
	}

	return cfg
}

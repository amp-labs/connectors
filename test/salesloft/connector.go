package salesloft

import (
	"context"
	"golang.org/x/oauth2"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesloft"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSalesloftConnector(ctx context.Context) *salesloft.Connector {
	filePath := credscanning.LoadPath(providers.Salesloft)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := salesloft.NewConnector(
		salesloft.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating Salesloft connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.salesloft.com/oauth/authorize",
			TokenURL:  "https://accounts.salesloft.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{},
	}

	return cfg
}

package attio

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetAttioConnector(ctx context.Context) *attio.Connector {
	filePath := credscanning.LoadPath(providers.Attio)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := attio.NewConnector(
		attio.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating Attio connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.attio.com/authorize",
			TokenURL:  "https://app.attio.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

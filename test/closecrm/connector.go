package closecrm

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/closecrm"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetCloseConnector(ctx context.Context) *closecrm.Connector {
	filePath := credscanning.LoadPath(providers.Close)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := closecrm.NewConnector(
		closecrm.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.close.com/oauth2/authorize",
			TokenURL:  "https://app.close.com/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"all.full_access",
			"offline_access",
		},
	}

	return &cfg
}

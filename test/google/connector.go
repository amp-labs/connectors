package google

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/google"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetGoogleConnector(ctx context.Context, moduleID common.ModuleID) *google.Connector {
	filePath := credscanning.LoadPath(providers.Google)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := google.NewConnector(
		google.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		google.WithModule(moduleID),
	)
	if err != nil {
		utils.Fail("error creating Google connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:  "https://oauth2.googleapis.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
		},
	}
}

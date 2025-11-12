package zoho

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetZohoConnector(ctx context.Context, module common.ModuleID) *zoho.Connector {
	filePath := credscanning.LoadPath(providers.Zoho)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := zoho.NewConnector(
		zoho.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		zoho.WithModule(module),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	fmt.Println("Module: ", module)

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.zoho.com/oauth/v2/auth",
			TokenURL:  "https://accounts.zoho.com/oauth/v2/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		Scopes: []string{
			"ZohoCRM.modules.ALL",
			"ZohoCRM.settings.ALL",
			"ZohoCRM.notifications.ALL",
		},
	}

	return &cfg
}

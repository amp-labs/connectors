package zoom

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetZoomConnector(ctx context.Context) *zoom.Connector {
	filePath := credscanning.LoadPath(providers.Zoom)

	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := zoom.NewConnector(zoom.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()))

	if err != nil {
		utils.Fail("error creating zoom connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://zoom.us/oauth/authorize",
			TokenURL:  "https://zoom.us/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

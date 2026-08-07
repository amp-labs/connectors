package zoominfo

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoominfo"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetZoomInfoConnector(ctx context.Context) *zoominfo.Connector {
	filePath := credscanning.LoadPath(providers.ZoomInfo)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := zoominfo.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating ZoomInfo connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://api.zoominfo.com/gtm/oauth/v1/authorize",
			TokenURL:  "https://api.zoominfo.com/gtm/oauth/v1/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

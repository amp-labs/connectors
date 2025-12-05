package dropboxsign

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/dropboxsign"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetDropboxSignConnector(ctx context.Context) *dropboxsign.Connector {
	filePath := credscanning.LoadPath(providers.DropboxSign)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating DropboxSign connector", "error", err)
	}

	conn, err := dropboxsign.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating DropboxSign connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.hellosign.com/oauth/authorize",
			TokenURL:  "https://app.hellosign.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

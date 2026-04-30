package goTo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/goTo"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

// GetGoToConnector builds a live GoTo Webinar connector using credentials
// loaded from go-to-creds.json (or wherever credscanning.LoadPath resolves).
//
// The organizer key is read from metadata.workspace in the creds file — set
// it to the value GoTo returns as `organizer_key` in its OAuth token response.
func GetGoToConnector(ctx context.Context) *goTo.Connector {
	filePath := credscanning.LoadPath(providers.GoTo)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := goTo.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
		Workspace:           reader.Get(credscanning.Fields.Workspace),
	})
	if err != nil {
		utils.Fail("error creating goTo connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://authentication.logmeininc.com/oauth/authorize",
			TokenURL:  "https://authentication.logmeininc.com/oauth/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}
}

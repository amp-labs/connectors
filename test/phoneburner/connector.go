package phoneburner

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetPhoneBurnerConnector(ctx context.Context) *phoneburner.Connector {
	filePath := credscanning.LoadPath(providers.PhoneBurner)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client := utils.NewOauth2Client(ctx, reader, getOAuthConfig())

	conn, err := phoneburner.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating PhoneBurner connector", "error", err)
	}

	return conn
}

func getOAuthConfig() func(*credscanning.ProviderCredentials) *oauth2.Config {
	return func(reader *credscanning.ProviderCredentials) *oauth2.Config {
		return &oauth2.Config{
			ClientID:     reader.Get(credscanning.Fields.ClientId),
			ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
			RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://www.phoneburner.com/oauth/authorize",
				TokenURL:  "https://www.phoneburner.com/oauth/accesstoken",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes: []string{},
		}
	}
}


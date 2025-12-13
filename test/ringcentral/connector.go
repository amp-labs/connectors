package ringcentral

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/ringcentral"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func NewConnector(ctx context.Context) (*ringcentral.Connector, error) {
	filePath := credscanning.LoadPath(providers.RingCentral)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return ringcentral.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://platform.ringcentral.com/restapi/oauth/authorize",
			TokenURL:  "https://platform.ringcentral.com/restapi/oauth/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		Scopes: []string{},
	}

	return &cfg
}

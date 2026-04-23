package talkdesk

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/talkdesk"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func NewConnector(ctx context.Context) *talkdesk.Connector {
	filePath := credscanning.LoadPath(providers.Talkdesk)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := talkdesk.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           "ampersand-dev",
		Metadata: map[string]string{
			"talkdesk_api_domain":   "api.talkdeskapp.com",
			"talkdesk_token_domain": "talkdeskid.com",
			"workspace":             "ampersand-dev",
		},
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.%s/oauth/authorize", "ampersand-dev", "talkdeskid.com"),
			TokenURL:  fmt.Sprintf("https://%s.%s/oauth/token", "ampersand-dev", "talkdeskid.com"),
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}

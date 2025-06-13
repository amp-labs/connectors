package zendeskchat

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendeskchat"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetConnector(ctx context.Context) *zendeskchat.Connector {
	filePath := credscanning.LoadPath(providers.ZendeskChat)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := zendeskchat.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           "d3v-ampersand",
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
			AuthURL:   fmt.Sprintf("https://%s.zendesk.com/oauth2/chat/authorizations/new", "d3v-ampersand"),
			TokenURL:  fmt.Sprintf("https://%s.zendesk.com/oauth2/chat/token", "d3v-ampersand"),
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
	}

	return &cfg
}

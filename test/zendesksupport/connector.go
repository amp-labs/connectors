package zendesksupport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	zendesksupport2 "github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetZendeskSupportConnector(ctx context.Context) *zendesksupport2.Connector {
	filePath := credscanning.LoadPath(providers.ZendeskSupport)
	reader := utils.MustCreateProvCredJSON(filePath, true, true)

	conn, err := zendesksupport2.NewConnector(
		zendesksupport2.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		zendesksupport2.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	workspace := reader.Get(credscanning.Fields.Workspace)

	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  fmt.Sprintf("https://%v.zendesk.com", workspace),
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%v.zendesk.com/oauth/authorizations/new", workspace),
			TokenURL:  fmt.Sprintf("https://%v.zendesk.com/oauth/tokens", workspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"read",
			"write",
		},
	}
}

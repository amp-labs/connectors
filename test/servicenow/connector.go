package servicenow

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/servicenow"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetServiceNowConnector(ctx context.Context) *servicenow.Connector {
	filePath := credscanning.LoadPath(providers.ServiceNow)
	reader := utils.MustCreateProvCredJSON(filePath, true, true)

	conn, err := servicenow.NewConnector(
		servicenow.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		servicenow.WithModule(servicenow.ModuleTable),
		servicenow.WithWorkspace("dev269415"),
	)
	if err != nil {
		utils.Fail("create servicenow connector", "error: ", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	workspace := reader.Get(credscanning.Fields.Workspace)

	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.service-now.com/oauth_auth.do", workspace),
			TokenURL:  fmt.Sprintf("https://%s.service-now.com/oauth_token.do", workspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

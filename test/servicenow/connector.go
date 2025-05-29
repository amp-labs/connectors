package servicenow

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/servicenow"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetServiceNowConnector(ctx context.Context) *servicenow.Connector {
	filePath := credscanning.LoadPath(providers.ServiceNow)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := servicenow.NewConnector(parameters.Connector{
		AuthenticatedClient: client,
		Workspace:           reader.Get(credscanning.Fields.Workspace),
	})
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

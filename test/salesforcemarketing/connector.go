package salesforcemarketing

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	salesforcemarketing "github.com/amp-labs/connectors/providers/salesforcemarketing"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetSalesforceMarketingConnector(ctx context.Context) *salesforcemarketing.Connector {
	filePath := credscanning.LoadPath(providers.SalesforceMarketing)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	options := []common.OAuthOption{
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(&oauth2.Token{
			AccessToken:  reader.Get(credscanning.Fields.AccessToken),
			RefreshToken: reader.Get(credscanning.Fields.RefreshToken),
		}),
	}

	client, err := common.NewOAuthHTTPClient(ctx, options...)

	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := salesforcemarketing.NewConnector(common.Parameters{
		AuthenticatedClient: client,
		Workspace:           reader.Get(credscanning.Fields.Workspace),
	})
	if err != nil {
		utils.Fail("error creating Salesforce Marketing Cloud connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			TokenURL:  "https://{{.workspace}}.auth.marketingcloudapis.com/v2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

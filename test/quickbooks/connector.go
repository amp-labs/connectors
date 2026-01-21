package quickbooks

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/quickbooks"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetQuickBooksConnector(ctx context.Context) *quickbooks.Connector {
	filePath := credscanning.LoadPath(providers.QuickBooks)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	client, err := common.NewOAuthHTTPClient(ctx,
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(reader)),
		common.WithOAuthToken(reader.GetOauthToken()),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := quickbooks.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata: map[string]string{
			"realmID": "9341455309256114", // QuickBooks Company ID, should be set dynamically
		},
	})
	if err != nil {
		utils.Fail("create quickbooks connector", "error: ", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://appcenter.intuit.com/connect/oauth2",
			TokenURL:  "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"com.intuit.quickbooks.accounting",
			"app-foundations.custom-field-definitions.read",
		},
		RedirectURL: "https://api.withampersand.com/callbacks/v1/oauth",
	}

	return cfg
}

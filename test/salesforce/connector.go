package salesforce

import (
	"context"
	"fmt"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
	"net/http"

	"github.com/amp-labs/connectors/salesforce"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetSalesforceConnector(ctx context.Context) *salesforce.Connector {
	reader := getSalesforceJSONReader()

	conn, err := salesforce.NewConnector(
		salesforce.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		salesforce.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	workspace := reader.Get(credscanning.Fields.Workspace)

	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", workspace),
			TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", workspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

func GetSalesforceAccessToken() string {
	reader := getSalesforceJSONReader()

	return reader.Get(credscanning.Fields.AccessToken)
}

func getSalesforceJSONReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.Salesforce)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	return reader
}

package salesforce

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	salesforce2 "github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetSalesforceConnector(ctx context.Context) *salesforce2.Connector {
	reader := getSalesforceJSONReader()

	conn, err := salesforce2.NewConnector(
		salesforce2.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		salesforce2.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
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
	reader := utils.MustCreateProvCredJSON(filePath, true, true)

	return reader
}

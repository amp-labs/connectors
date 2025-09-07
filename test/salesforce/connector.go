package salesforce

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var fieldBusinessUnitID = credscanning.Field{
	Name:      "businessUnitId",
	PathJSON:  "metadata.businessUnitId",
	SuffixENV: "BUSINESS_UNIT_ID",
}

func GetSalesforceConnector(ctx context.Context) *salesforce.Connector {
	return getSalesforceConnector(ctx, providers.ModuleSalesforceCRM)
}

func GetSalesforceAccountEngagementConnector(ctx context.Context) *salesforce.Connector {
	return getSalesforceConnector(ctx, providers.ModuleSalesforceAccountEngagementDemo)
}

func getSalesforceConnector(ctx context.Context, module common.ModuleID) *salesforce.Connector {
	reader := getSalesforceJSONReader()

	conn, err := salesforce.NewConnector(
		salesforce.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		salesforce.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
		salesforce.WithModule(module),
		salesforce.WithMetadata(map[string]string{
			"businessUnitId": reader.Get(fieldBusinessUnitID),
		}),
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

func GetSalesforceAccessToken() common.AuthToken {
	reader := getSalesforceJSONReader()

	return common.AuthToken(reader.Get(credscanning.Fields.AccessToken))
}

func getSalesforceJSONReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.Salesforce)
	reader := utils.MustCreateProvCredJSON(filePath, true,
		fieldBusinessUnitID,
	)

	return reader
}

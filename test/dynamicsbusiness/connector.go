package dynamicsbusiness

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/dynamicsbusiness"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

// nolint:gochecknoglobals
var (
	fieldCompanyID = credscanning.Field{
		Name:      "region",
		PathJSON:  "metadata.companyId",
		SuffixENV: "REGION",
	}
	fieldEnvironmentName = credscanning.Field{
		Name:      "environmentName",
		PathJSON:  "metadata.environmentName",
		SuffixENV: "ENVIRONMENT_NAME",
	}
)

func GetDynamicsBusinessCentralConnector(ctx context.Context) *dynamicsbusiness.Connector {
	filePath := credscanning.LoadPath(providers.DynamicsBusinessCentral)
	reader := utils.MustCreateProvCredJSON(filePath, true,
		fieldCompanyID, fieldEnvironmentName,
	)

	conn, err := dynamicsbusiness.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
			Metadata: map[string]string{
				"companyId":       reader.Get(fieldCompanyID),
				"environmentName": reader.Get(fieldEnvironmentName),
			},
		},
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
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://login.microsoftonline.com/%v/oauth2/v2.0/authorize", workspace),
			TokenURL:  fmt.Sprintf("https://login.microsoftonline.com/%v/oauth2/v2.0/token", workspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"https://api.businesscentral.dynamics.com/user_impersonation",
			"offline_access",
		},
	}
}

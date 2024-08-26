package dynamicscrm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/dynamicscrm"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetMSDynamics365CRMConnector(ctx context.Context) *dynamicscrm.Connector {
	filePath := credscanning.LoadPath(providers.DynamicsCRM)
	reader := utils.MustCreateProvCredJSON(filePath, true, true)

	conn, err := dynamicscrm.NewConnector(
		dynamicscrm.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		dynamicscrm.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
	)

	if err != nil {
		utils.Fail("error creating microsoft CRM connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			fmt.Sprintf("https://%v.crm.dynamics.com/user_impersonation",
				reader.Get(credscanning.Fields.Workspace)),
			"offline_access",
		},
	}
}

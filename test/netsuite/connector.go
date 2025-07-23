package netsuite

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/netsuite"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetNetsuiteConnector(ctx context.Context) *netsuite.Connector {
	filePath := credscanning.LoadPath(providers.Netsuite)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	options := []common.OAuthOption{
		common.WithOAuthClient(http.DefaultClient),
		common.WithOAuthConfig(getConfig(nil)),
		common.WithOAuthToken(&oauth2.Token{
			AccessToken:  reader.Get(credscanning.Fields.AccessToken),
			RefreshToken: reader.Get(credscanning.Fields.RefreshToken),
		}),
	}

	client, err := common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		panic(err)
	}

	conn, err := netsuite.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           reader.Get(credscanning.Fields.Workspace),
	})
	if err != nil {
		utils.Fail("error creating netsuite connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "https://api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://{{.workspace}}.app.netsuite.com/app/login/oauth2/authorize.nl",
			TokenURL:  "https://{{.workspace}}.suitetalk.api.netsuite.com/services/rest/auth/oauth2/v1/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"email",
			"rest_webservices",
			"suite_analytics",
		},
	}

	return cfg
}

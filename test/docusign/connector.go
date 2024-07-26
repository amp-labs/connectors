package docusign

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/docusign"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetDocusignConnector(ctx context.Context) *docusign.Connector {
	filePath := credscanning.LoadPath(providers.Docusign)
	reader := utils.MustCreateProvCredJSON(filePath, true, false)

	conn, err := docusign.NewConnector(
		docusign.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		docusign.WithMetadata(map[string]string{
			// This value can be obtained by following this API reference.
			// https://developers.docusign.com/platform/auth/reference/user-info
			"server": "na3",
		}),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://account.docusign.com/oauth/auth",
			TokenURL:  "https://account.docusign.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{},
	}
}

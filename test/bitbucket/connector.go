package bitbucket

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/bitbucket"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var fieldWorkspace = credscanning.Field{
	Name:      "workspace",
	PathJSON:  "metadata.workspace",
	SuffixENV: "WORKSPACE",
}

func GetConnector(ctx context.Context) *bitbucket.Connector {
	filePath := credscanning.LoadPath(providers.Bitbucket)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldWorkspace)

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

	mp := make(map[string]string)
	mp["workspace"] = reader.Get(fieldWorkspace)

	conn, err := bitbucket.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Metadata:            mp,
		Workspace:           reader.Get(fieldWorkspace),
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://bitbucket.org/site/oauth2/authorize",
			TokenURL:  "https://bitbucket.org/site/oauth2/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

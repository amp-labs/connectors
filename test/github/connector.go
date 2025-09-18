package github

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	github "github.com/amp-labs/connectors/providers/github"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

func GetGithubConnector(ctx context.Context) *github.Connector {
	filePath := credscanning.LoadPath(providers.Github)
	reader := utils.MustCreateProvCredJSON(filePath, true)

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

	conn, err := github.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: client,
		},
	)
	if err != nil {
		utils.Fail("error creating github connector", "error", err)
	}

	return conn
}

func getConfig(reader *credscanning.ProviderCredentials) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID:     reader.Get(credscanning.Fields.ClientId),
		ClientSecret: reader.Get(credscanning.Fields.ClientSecret),
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://github.com/login/oauth/authorize",
			TokenURL:  "https://github.com/login/oauth/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

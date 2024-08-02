package atlassian

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/atlassian"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

const cloudId = "ebc887b2-7e61-4059-ab35-71f15cc16e12"

func GetAtlassianConnector(ctx context.Context) *atlassian.Connector {
	filePath := credscanning.LoadPath(providers.Atlassian)
	reader := utils.MustCreateProvCredJSON(filePath, true, true)

	conn, err := atlassian.NewConnector(
		atlassian.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		atlassian.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
		atlassian.WithModule(atlassian.ModuleJira),
		atlassian.WithMetadata(map[string]string{
			// This value can be obtained by following this API reference.
			// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
			"cloudId": cloudId,
		}),
	)
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
			AuthURL:   "https://auth.atlassian.com/authorize",
			TokenURL:  "https://auth.atlassian.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"offline_access",
			"read:jira-user",
			"read:jira-work",
			"write:jira-work",
			"manage:jira-project",
			"manage:jira-configuration",
			"manage:jira-webhook",
		},
	}

	return cfg
}

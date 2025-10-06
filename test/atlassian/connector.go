package atlassian

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/atlassian"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

const cloudId = "35745fff-f0de-466c-b08e-a63f69888611"

func GetAtlassianConnector(ctx context.Context) *atlassian.Connector {
	filePath := credscanning.LoadPath(providers.Atlassian)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	conn, err := atlassian.NewConnector(
		atlassian.WithClient(ctx, http.DefaultClient, getConfig(reader), reader.GetOauthToken()),
		atlassian.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
		atlassian.WithModule(providers.ModuleAtlassianJira),
		atlassian.WithMetadata(map[string]string{
			// This value can be obtained by following this API reference.
			// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
			// Another simplest solution is to run `connectors/test/atlassian/auth-metadata/main.go` script.
			"cloudId": cloudId,
		}),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func GetAtlassianConnectConnector(ctx context.Context, claims map[string]any) *atlassian.Connector {
	filePath := credscanning.LoadPath(providers.Atlassian)
	reader := utils.MustCreateProvCredJSON(filePath, true)

	opts := []common.HeaderAuthClientOption{
		common.WithHeaderClient(http.DefaultClient),
		common.WithDynamicHeaders(atlassian.JwtTokenGenerator(claims, reader.Get(credscanning.Fields.Secret))),
	}

	client, err := common.NewHeaderAuthHTTPClient(ctx, opts...)
	if err != nil {
		panic(err)
	}

	conn, err := atlassian.NewConnector(
		atlassian.WithAuthenticatedClient(client),
		atlassian.WithWorkspace(reader.Get(credscanning.Fields.Workspace)),
		atlassian.WithModule(providers.ModuleAtlassianJiraConnect),
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
		panic(err)
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

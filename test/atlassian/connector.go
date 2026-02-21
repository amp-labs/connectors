package atlassian

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/atlassian"
	"github.com/amp-labs/connectors/test/utils"
	"golang.org/x/oauth2"
)

var fieldCloudID = credscanning.Field{
	Name:      "cloudId",
	PathJSON:  "metadata.cloudId",
	SuffixENV: "CLOUD_ID",
}

func GetAtlassianConnector(ctx context.Context) *atlassian.Connector {
	return makeAtlassianConnector(ctx, providers.ModuleAtlassianJira)
}

func makeAtlassianConnector(ctx context.Context, module common.ModuleID) *atlassian.Connector {
	filePath := credscanning.LoadPath(providers.Atlassian)
	reader := utils.MustCreateProvCredJSON(filePath, true, fieldCloudID)

	conn, err := atlassian.NewConnector(
		common.ConnectorParams{
			Module:              module,
			AuthenticatedClient: utils.NewOauth2Client(ctx, reader, getConfig),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
			Metadata: map[string]string{
				// This value can be obtained by following this API reference.
				// https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3-1-get-the-cloudid-for-your-site
				// Another simplest solution is to run `connectors/test/atlassian/auth-metadata/main.go` script.
				"cloudId": reader.Get(fieldCloudID),
			},
		},
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

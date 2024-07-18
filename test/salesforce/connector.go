package salesforce

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/salesforce"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetSalesforceConnector(ctx context.Context, filePath string) *salesforce.Connector {
	registry := utils.NewCredentialsRegistry()

	readers := []utils.Reader{
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.clientId",
			CredKey:  utils.ClientId,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.clientSecret",
			CredKey:  utils.ClientSecret,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.accessToken",
			CredKey:  utils.AccessToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.refreshToken",
			CredKey:  utils.RefreshToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.provider",
			CredKey:  utils.Provider,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.substitutions.workspace",
			CredKey:  utils.WorkspaceRef,
		},
	}
	_ = registry.AddReaders(readers...)

	cfg := utils.SalesforceOAuthConfigFromRegistry(registry)
	tok := utils.SalesforceOauthTokenFromRegistry(registry)
	workspace := registry.MustString(utils.WorkspaceRef)

	conn, err := salesforce.NewConnector(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithWorkspace(workspace),
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}

func GetSalesforceAccessToken(filePath string) string {
	registry := utils.NewCredentialsRegistry()

	_ = registry.AddReaders(&utils.JSONReader{
		FilePath: filePath,
		JSONPath: "$.accessToken",
		CredKey:  utils.AccessToken,
	})

	return registry.MustString(utils.AccessToken)
}

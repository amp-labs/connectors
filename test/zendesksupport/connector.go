package zendesksupport

import (
	"context"
	"net/http"

	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
	"github.com/amp-labs/connectors/zendesksupport"
)

func GetZendeskSupportConnector(ctx context.Context, filePath string) *zendesksupport.Connector {
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

	cfg := utils.ZendeskSupportConfigFromRegistry(registry)
	tok := utils.ZendeskSupportTokenFromRegistry(registry)
	workspace := registry.MustString(utils.WorkspaceRef)

	conn, err := zendesksupport.NewConnector(
		zendesksupport.WithClient(ctx, http.DefaultClient, cfg, tok),
		zendesksupport.WithWorkspace(workspace),
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}

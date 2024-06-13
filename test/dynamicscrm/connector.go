package dynamicscrm

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/dynamicscrm"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetMSDynamics365CRMConnector(ctx context.Context, filePath string) *dynamicscrm.Connector {
	registry := utils.NewCredentialsRegistry()

	readers := []utils.Reader{
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_ID",
			CredKey:  utils.ClientId,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_SECRET",
			CredKey:  utils.ClientSecret,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.ACCESS_TOKEN",
			CredKey:  utils.AccessToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.REFRESH_TOKEN",
			CredKey:  utils.RefreshToken,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.PROVIDER",
			CredKey:  utils.Provider,
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.WORKSPACE",
			CredKey:  utils.WorkspaceRef,
		},
	}
	_ = registry.AddReaders(readers...)

	cfg := utils.MSDynamics365CRMConfigFromRegistry(registry)
	tok := utils.MSDynamics365CRMTokenFromRegistry(registry)
	workspace := registry.MustString(utils.WorkspaceRef)

	conn, err := dynamicscrm.NewConnector(
		dynamicscrm.WithClient(ctx, http.DefaultClient, cfg, tok),
		dynamicscrm.WithWorkspace(workspace),
	)
	if err != nil {
		testUtils.Fail("error creating microsoft CRM connector", "error", err)
	}

	return conn
}

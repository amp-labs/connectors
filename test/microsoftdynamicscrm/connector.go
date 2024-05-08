package microsoftdynamicscrm

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/microsoftdynamicscrm"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetMSDynamics365CRMConnector(ctx context.Context, filePath string) *microsoftdynamicscrm.Connector {
	registry := utils.NewCredentialsRegistry()

	readers := []utils.Reader{
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_ID",
			CredKey:  "clientId",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_SECRET",
			CredKey:  "clientSecret",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.ACCESS_TOKEN",
			CredKey:  "accessToken",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.REFRESH_TOKEN",
			CredKey:  "refreshToken",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.PROVIDER",
			CredKey:  "provider",
		},
	}
	_ = registry.AddReaders(readers...)

	cfg := utils.MSDynamics365CRMConfigFromRegistry(registry)
	tok := utils.MSDynamics365CRMTokenFromRegistry(registry)

	conn, err := connectors.MSDynamicsCRM(
		microsoftdynamicscrm.WithClient(ctx, http.DefaultClient, cfg, tok),
		microsoftdynamicscrm.WithWorkspace(utils.MSDynamics365CRMWorkspace),
		microsoftdynamicscrm.WithModule(microsoftdynamicscrm.DefaultModuleCRM),
	)
	if err != nil {
		testUtils.Fail("error creating microsoft CRM connector", "error", err)
	}

	return conn
}

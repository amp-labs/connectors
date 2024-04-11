package msdsales

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/msdsales"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetMSDynamics365SalesConnector(ctx context.Context, filePath string) *msdsales.Connector {
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

	cfg := utils.MSDynamics365SalesConfigFromRegistry(registry)
	tok := utils.MSDynamics365SalesTokenFromRegistry(registry)

	conn, err := connectors.MSDynamicsSales(
		msdsales.WithClient(ctx, http.DefaultClient, cfg, tok),
		msdsales.WithWorkspace(utils.MSDynamics365SalesWorkspace),
		msdsales.WithModule(msdsales.DefaultModuleCRM),
	)
	if err != nil {
		testUtils.Fail("error creating microsoft sales connector", "error", err)
	}

	return conn
}

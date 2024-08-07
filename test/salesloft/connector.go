package salesloft

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/salesloft"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetSalesloftConnector(ctx context.Context, filePath string) *salesloft.Connector {
	registry := scanning.NewRegistry()

	readers := []scanning.Reader{
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_ID",
			KeyName:  "clientId",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_SECRET",
			KeyName:  "clientSecret",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.ACCESS_TOKEN",
			KeyName:  "accessToken",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.REFRESH_TOKEN",
			KeyName:  "refreshToken",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.PROVIDER",
			KeyName:  "provider",
		},
	}
	_ = registry.AddReaders(readers...)

	cfg := utils.SalesloftConfigFromRegistry(registry)
	tok := utils.SalesloftTokenFromRegistry(registry)

	conn, err := salesloft.NewConnector(
		salesloft.WithClient(ctx, http.DefaultClient, cfg, tok),
	)
	if err != nil {
		testUtils.Fail("error creating Salesloft connector", "error", err)
	}

	return conn
}

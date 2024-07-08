package salesloft

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/credsregistry"
	"github.com/amp-labs/connectors/salesloft"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetSalesloftConnector(ctx context.Context, filePath string) *salesloft.Connector {
	registry := credsregistry.NewCredentialsRegistry()

	readers := []credsregistry.Reader{
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_ID",
			CredKey:  "clientId",
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.CLIENT_SECRET",
			CredKey:  "clientSecret",
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.ACCESS_TOKEN",
			CredKey:  "accessToken",
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.REFRESH_TOKEN",
			CredKey:  "refreshToken",
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.PROVIDER",
			CredKey:  "provider",
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

package hubspot

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/credsregistry"
	"github.com/amp-labs/connectors/hubspot"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

// GetHubspotConnector returns a Hubspot connector.
func GetHubspotConnector(ctx context.Context, filePath string) *hubspot.Connector {
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
			JSONPath: "$.REFRESH_TOKEN",
			CredKey:  "refreshToken",
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.ACCESS_TOKEN",
			CredKey:  "accessToken",
		},
		&credsregistry.JSONReader{
			FilePath: filePath,
			JSONPath: "$.PROVIDER",
			CredKey:  "provider",
		},
	}
	registry.AddReaders(readers...)

	cfg := utils.HubspotOAuthConfigFromRegistry(registry)
	tok := utils.HubspotOauthTokenFromRegistry(registry)

	conn, err := hubspot.NewConnector(
		hubspot.WithClient(ctx, http.DefaultClient, cfg, tok),
		hubspot.WithModule(hubspot.ModuleCRM))
	if err != nil {
		testUtils.Fail("error creating hubspot connector", "error", err)
	}

	return conn
}

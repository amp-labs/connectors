package hubspot

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/providers/hubspot"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

// GetHubspotConnector returns a Hubspot connector.
func GetHubspotConnector(ctx context.Context, filePath string) *hubspot.Connector {
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
			JSONPath: "$.REFRESH_TOKEN",
			CredKey:  "refreshToken",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$.ACCESS_TOKEN",
			CredKey:  "accessToken",
		},
		&utils.JSONReader{
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

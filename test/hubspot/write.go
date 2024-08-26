package hubspot

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/providers/hubspot"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

// GetHubspotConnector returns a Hubspot connector.
func GetHubspotConnector(ctx context.Context, filePath string) *hubspot.Connector {
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
			JSONPath: "$.REFRESH_TOKEN",
			KeyName:  "refreshToken",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.ACCESS_TOKEN",
			KeyName:  "accessToken",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$.PROVIDER",
			KeyName:  "provider",
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

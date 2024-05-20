package intercom

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/intercom"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetIntercomConnector(ctx context.Context, filePath string) *intercom.Connector {
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
	}
	_ = registry.AddReaders(readers...)

	cfg := utils.IntercomConfigFromRegistry(registry)
	tok := utils.IntercomTokenFromRegistry(registry)

	conn, err := connectors.Intercom(
		intercom.WithClient(ctx, http.DefaultClient, cfg, tok),
		intercom.WithModule(intercom.DefaultModule),
	)
	if err != nil {
		testUtils.Fail("error creating Intercom connector", "error", err)
	}

	return conn
}

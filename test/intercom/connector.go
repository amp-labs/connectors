package intercom

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/intercom"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetIntercomConnector(ctx context.Context, filePath string) *intercom.Connector {
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
	}
	_ = registry.AddReaders(readers...)

	cfg := utils.IntercomConfigFromRegistry(registry)
	tok := utils.IntercomTokenFromRegistry(registry)

	conn, err := intercom.NewConnector(
		intercom.WithClient(ctx, http.DefaultClient, cfg, tok),
	)
	if err != nil {
		testUtils.Fail("error creating Intercom connector", "error", err)
	}

	return conn
}

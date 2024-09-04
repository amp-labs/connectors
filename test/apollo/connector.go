package apollo

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/apollo"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

func GetApolloConnector(ctx context.Context, filePath string) *apollo.Connector {
	registry := scanning.NewRegistry()

	readers := []scanning.Reader{
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$['apiKey']",
			KeyName:  "apiKey",
		},
	}
	_ = registry.AddReaders(readers...)

	apiKey := utils.ApolloAPIKeyFromRegistry(registry)

	conn, err := apollo.NewConnector(
		apollo.WithClient(ctx, http.DefaultClient, apiKey),
	)
	if err != nil {
		testUtils.Fail("error creating Apollo connector", "error", err)
	}

	return conn
}

func GetApolloJSONReader() *credscanning.ProviderCredentials {
	filePath := credscanning.LoadPath(providers.Apollo)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, false)

	return reader
}

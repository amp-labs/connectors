package apollo

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/apollo"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetApolloConnector(ctx context.Context) *apollo.Connector {
	filePath := credscanning.LoadPath(providers.Apollo)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := apollo.NewConnector(
		apollo.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		testUtils.Fail("error creating Apollo connector", "error", err)
	}

	return conn
}

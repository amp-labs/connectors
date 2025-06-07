package front

import (
	"context"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/front"
	"github.com/amp-labs/connectors/test/utils"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetFrontConnector(ctx context.Context) *front.Connector {
	filePath := credscanning.LoadPath(providers.Front)
	reader := testUtils.MustCreateProvCredJSON(filePath, false)

	conn, err := front.NewConnector(parameters.Connector{
		AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Front),
	})
	if err != nil {
		testUtils.Fail("error creating Front App connector", "error", err)
	}

	return conn
}

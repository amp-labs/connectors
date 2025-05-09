package fireflies

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/fireflies"
	"github.com/amp-labs/connectors/test/utils"
)

func GetFirefliesConnector(ctx context.Context) *fireflies.Connector {
	filePath := credscanning.LoadPath(providers.Fireflies)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Fireflies)

	conn, err := fireflies.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Fireflies App connector", "error", err)
	}

	return conn
}

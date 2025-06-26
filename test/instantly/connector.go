package instantly

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/instantly"
	"github.com/amp-labs/connectors/test/utils"
)

func GetInstantlyConnector(ctx context.Context) *instantly.Connector {
	filePath := credscanning.LoadPath(providers.InstantlyAI)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.InstantlyAI)

	conn, err := instantly.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Instantly App connector", "error", err)
	}

	return conn
}

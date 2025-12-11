package paddle

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/paddle"
	"github.com/amp-labs/connectors/test/utils"
)

func GetPaddleConnector(ctx context.Context) *paddle.Connector {
	filePath := credscanning.LoadPath(providers.Paddle)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Paddle)

	conn, err := paddle.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Paddle connector", "error", err)
	}

	return conn
}

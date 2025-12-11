package aircall

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aircall"
	"github.com/amp-labs/connectors/test/utils"
)

func GetAircallConnector(ctx context.Context) *aircall.Connector {
	filePath := credscanning.LoadPath(providers.Aircall)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Aircall)

	conn, err := aircall.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating aircall app connector", "error", err)
	}

	return conn
}

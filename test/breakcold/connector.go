package breakcold

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/breakcold"
	"github.com/amp-labs/connectors/test/utils"
)

func GetBreakcoldConnector(ctx context.Context) *breakcold.Connector {
	filePath := credscanning.LoadPath(providers.Breakcold)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Breakcold)

	conn, err := breakcold.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Fireflies App connector", "error", err)
	}

	return conn
}

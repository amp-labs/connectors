package salesfinity

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesfinity"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *salesfinity.Connector {
	filePath := credscanning.LoadPath(providers.Salesfinity)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Salesfinity)

	conn, err := salesfinity.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

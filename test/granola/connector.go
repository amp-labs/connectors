package granola

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/granola"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *granola.Connector {
	filePath := credscanning.LoadPath(providers.Granola)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Granola)

	conn, err := granola.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

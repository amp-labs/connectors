package salesflare

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesflare"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSalesflareConnector(ctx context.Context) *salesflare.Connector {
	filePath := credscanning.LoadPath(providers.Salesflare)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := salesflare.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Salesflare),
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

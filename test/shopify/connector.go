package shopify

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/shopify"
	"github.com/amp-labs/connectors/test/utils"
)

func GetShopifyConnector(ctx context.Context) *shopify.Connector {
	filePath := credscanning.LoadPath(providers.Shopify)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	workspace := reader.Get(credscanning.Fields.Workspace)
	client := utils.NewAPIKeyClient(ctx, reader, providers.Shopify)

	conn, err := shopify.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           workspace,
	})
	if err != nil {
		utils.Fail("error creating Shopify connector", "error", err)
	}

	return conn
}

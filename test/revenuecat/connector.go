package revenuecat

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
)

func GetRevenueCatConnector(ctx context.Context) *revenuecat.Connector {
	filePath := credscanning.LoadPath(providers.RevenueCat)

	// Load credentials with metadata fields
	reader := utils.MustCreateProvCredJSON(filePath, false)

	params := common.ConnectorParams{
		AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.RevenueCat),
	}

	conn, err := revenuecat.NewConnector(params)
	if err != nil {
		utils.Fail("error while creating revenuecat connector", "error", err)
	}

	return conn
}

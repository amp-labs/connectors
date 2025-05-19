package insightly

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/insightly"
	"github.com/amp-labs/connectors/test/utils"
)

func GetInsightlyConnector(ctx context.Context) *insightly.Connector {
	filePath := credscanning.LoadPath(providers.Insightly)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := insightly.NewConnector(
		common.Parameters{
			AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

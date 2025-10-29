package recurly

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/recurly"
	"github.com/amp-labs/connectors/test/utils"
)

func GetRecurlyConnector(ctx context.Context) *recurly.Connector {
	filePath := credscanning.LoadPath(providers.Recurly)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := recurly.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

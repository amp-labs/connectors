package monaco

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/monaco"
	"github.com/amp-labs/connectors/test/utils"
)

func GetMonacoConnector(ctx context.Context) *monaco.Connector {
	filePath := credscanning.LoadPath(providers.Monaco)

	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := monaco.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Monaco),
		},
	)
	if err != nil {
		utils.Fail("error creating Monaco connector", "error", err)
	}

	return conn
}

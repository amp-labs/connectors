package connectWise

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/connectwise"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnectWiseConnector(ctx context.Context) *connectwise.Connector {
	filePath := credscanning.LoadPath(providers.ConnectWise)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := connectwise.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

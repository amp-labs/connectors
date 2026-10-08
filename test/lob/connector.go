package lob

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/lob"
	"github.com/amp-labs/connectors/test/utils"
)

func GetLobConnector(ctx context.Context) *lob.Connector {
	filePath := credscanning.LoadPath(providers.Lob)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := lob.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

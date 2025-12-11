package outplay

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/outplay"
	"github.com/amp-labs/connectors/test/utils"
)

func GetOutplayConnector(ctx context.Context) *outplay.Connector {
	filePath := credscanning.LoadPath(providers.Outplay)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := outplay.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

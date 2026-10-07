package reply

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/reply"
	"github.com/amp-labs/connectors/test/utils"
)

func GetReplyConnector(ctx context.Context) *reply.Connector {
	filePath := credscanning.LoadPath(providers.Reply)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Reply)

	conn, err := reply.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating reply connector", "error", err)
	}

	return conn
}

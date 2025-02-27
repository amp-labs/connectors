package blueshift

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/blueshift"
	"github.com/amp-labs/connectors/test/utils"
)

func GetBlueshiftConnector(ctx context.Context) *blueshift.Connector {
	filePath := credscanning.LoadPath(providers.Blueshift)

	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client, err := common.NewBasicAuthHTTPClient(ctx, reader.Get(credscanning.Fields.Username), reader.Get(credscanning.Fields.Password))

	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := blueshift.NewConnector(
		common.Parameters{
			AuthenticatedClient: client,
		},
	)
	if err != nil {
		utils.Fail("error creating asana connector", "error", err)
	}

	return conn
}

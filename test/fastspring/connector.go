package fastspring

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/fastspring"
	"github.com/amp-labs/connectors/test/utils"
)

func GetFastSpringConnector(ctx context.Context) *fastspring.Connector {
	filePath := credscanning.LoadPath(providers.FastSpring)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := fastspring.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
		},
	)
	if err != nil {
		utils.Fail("error creating FastSpring connector", "error", err)
	}

	return conn
}

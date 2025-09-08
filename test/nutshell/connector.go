package nutshell

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/nutshell"
	"github.com/amp-labs/connectors/test/utils"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetNutshellConnector(ctx context.Context) *nutshell.Connector {
	filePath := credscanning.LoadPath(providers.Nutshell)
	reader := testUtils.MustCreateProvCredJSON(filePath, false)

	conn, err := nutshell.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
		},
	)
	if err != nil {
		testUtils.Fail("error creating connector", "error", err)
	}

	return conn
}

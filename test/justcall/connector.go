package justcall

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/justcall"
	"github.com/amp-labs/connectors/test/utils"
)

func GetJustCallConnector(ctx context.Context) *justcall.Connector {
	filePath := credscanning.LoadPath(providers.JustCall)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.JustCall)

	conn, err := justcall.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating justcall connector", "error", err)
	}

	return conn
}

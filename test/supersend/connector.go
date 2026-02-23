package supersend

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/supersend"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSuperSendConnector(ctx context.Context) *supersend.Connector {
	filePath := credscanning.LoadPath(providers.SuperSend)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.SuperSend)

	conn, err := supersend.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating superSend connector", "error", err)
	}

	return conn
}

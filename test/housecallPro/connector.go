package housecallpro

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/housecallPro"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *housecallpro.Connector {
	filePath := credscanning.LoadPath(providers.HousecallPro)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.HousecallPro)

	conn, err := housecallpro.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

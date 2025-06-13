package avoma

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/avoma"
	"github.com/amp-labs/connectors/test/utils"
)

func GetAvomaConnector(ctx context.Context) *avoma.Connector {
	filePath := credscanning.LoadPath(providers.Avoma)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Avoma)

	conn, err := avoma.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating avoma app connector", "error", err)
	}

	return conn
}

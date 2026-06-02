package acculynx

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

func GetAccuLynxConnector(ctx context.Context) *acculynx.Connector {
	filePath := credscanning.LoadPath(providers.AccuLynx)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.AccuLynx)

	conn, err := acculynx.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating acculynx connector", "error", err)
	}

	return conn
}

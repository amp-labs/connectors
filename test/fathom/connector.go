package fathom

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/fathom"
	"github.com/amp-labs/connectors/test/utils"
)

func GetFathomConnector(ctx context.Context) *fathom.Connector {
	filePath := credscanning.LoadPath(providers.Fathom)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Fathom)

	conn, err := fathom.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Fathom App connector", "error", err)
	}

	return conn
}

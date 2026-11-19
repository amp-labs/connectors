package auphan

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/auphan"
	"github.com/amp-labs/connectors/test/utils"
)

func GetAuphanConnector(ctx context.Context) *auphan.Connector {
	filePath := credscanning.LoadPath(providers.Auphan)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := auphan.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Auphan),
		Workspace:           reader.Get(credscanning.Fields.Workspace),
	})
	if err != nil {
		utils.Fail("error creating Auphan connector", "error", err)
	}

	return conn
}

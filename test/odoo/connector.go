package odoo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/odoo"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *odoo.Connector {
	filePath := credscanning.LoadPath(providers.Odoo)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.Odoo)

	workspace := reader.Get(credscanning.Fields.Workspace)
	if workspace == "" {
		utils.Fail("missing metadata.workspace in creds")
	}

	conn, err := odoo.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           workspace,
	})
	if err != nil {
		utils.Fail("error creating Odoo connector", "error", err)
	}

	return conn
}

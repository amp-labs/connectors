package chargeover

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/chargeover"
	"github.com/amp-labs/connectors/test/utils"
)

func NewConnector(ctx context.Context) *chargeover.Connector {
	filePath := credscanning.LoadPath(providers.ChargeOver)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewBasicAuthHTTPClient(ctx, reader.Get(credscanning.Fields.Username), reader.Get(credscanning.Fields.Password))
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := chargeover.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: client,
			Workspace:           reader.Get(credscanning.Fields.Workspace),
		},
	)
	if err != nil {
		utils.Fail("error creating chargeover connector", "error", err)
	}

	return conn
}

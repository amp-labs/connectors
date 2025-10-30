package chargebee

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/chargebee"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetChargebeeConnector(ctx context.Context) *chargebee.Connector {
	filePath := credscanning.LoadPath(providers.Chargebee)
	reader := testUtils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewBasicAuthHTTPClient(ctx, reader.Get(credscanning.Fields.Username), reader.Get(credscanning.Fields.Password))
	if err != nil {
		testUtils.Fail(err.Error())
	}

	conn, err := chargebee.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: client,
			Workspace:           reader.Get(credscanning.Fields.Workspace),
		},
	)
	if err != nil {
		testUtils.Fail("error creating Chargebee connector", "error", err)
	}

	return conn
}

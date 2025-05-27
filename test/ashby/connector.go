package ashby

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/ashby"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetAshbyConnector(ctx context.Context) *ashby.Connector {
	filePath := credscanning.LoadPath(providers.Ashby)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, false)

	client, err := common.NewBasicAuthHTTPClient(ctx, reader.Get(credscanning.Fields.Username), reader.Get(credscanning.Fields.Password))

	if err != nil {
		testUtils.Fail(err.Error())
	}

	conn, err := ashby.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: client,
		},
	)
	if err != nil {
		testUtils.Fail("error creating ashby connector", "error", err)
	}

	return conn
}

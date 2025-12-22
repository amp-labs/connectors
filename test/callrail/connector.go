package callrail

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/callrail"
	"github.com/amp-labs/connectors/test/utils"
)

func NewConnector(ctx context.Context) *callrail.Connector {
	filePath := credscanning.LoadPath(providers.CallRail)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", "Token token="+reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		return nil
	}

	conn, err := callrail.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: client,
		},
	)
	if err != nil {
		utils.Fail("error creating calendly connector", "error", err)
	}

	return conn
}

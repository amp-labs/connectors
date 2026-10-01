package wealthbox

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/wealthbox"
	"github.com/amp-labs/connectors/test/utils"
)

func GetWealthboxConnector(ctx context.Context) *wealthbox.Connector {
	filePath := credscanning.LoadPath(providers.Wealthbox)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "ACCESS_TOKEN", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := wealthbox.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

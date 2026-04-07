package freshchat

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/freshchat"
	"github.com/amp-labs/connectors/test/utils"
)

func NewConnector(ctx context.Context) *freshchat.Connector {
	filePath := credscanning.LoadPath(providers.Freshchat)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", "Bearer "+reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := freshchat.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client, Workspace: "marsinc-team-d57fdeb9aafa7f517755548"},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

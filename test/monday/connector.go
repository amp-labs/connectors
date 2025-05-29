package monday

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/monday"
	"github.com/amp-labs/connectors/test/utils"
)

func GetMondayConnector(ctx context.Context) *monday.Connector {
	filePath := credscanning.LoadPath(providers.Monday)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", reader.Get(credscanning.Fields.ApiKey))

	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := monday.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

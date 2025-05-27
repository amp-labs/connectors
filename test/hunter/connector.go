package hunter

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hunter"
	"github.com/amp-labs/connectors/test/utils"
)

func GetHunterConnector(ctx context.Context) *hunter.Connector {
	filePath := credscanning.LoadPath(providers.Hunter)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client, err := common.NewApiKeyQueryParamAuthHTTPClient(ctx, "api_key", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := hunter.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

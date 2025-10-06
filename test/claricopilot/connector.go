package claricopilot

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/claricopilot"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *claricopilot.Connector {
	filePath := credscanning.LoadPath(providers.ClariCopilot)
	reader := utils.MustCreateProvCredJSON(filePath, false, credscanning.Fields.ApiKey, credscanning.Fields.Secret)

	apiKey := reader.Get(credscanning.Fields.ApiKey)
	apiSecret := reader.Get(credscanning.Fields.Secret)

	client, err := common.NewCustomAuthHTTPClient(ctx, common.WithCustomHeaders(common.Header{Key: "X-Api-Key", Value: apiKey}, common.Header{Key: "X-Api-Password", Value: apiSecret}))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := claricopilot.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)

	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

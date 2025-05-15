package instantlyai

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/instantlyai"
	"github.com/amp-labs/connectors/test/utils"
)

func GetInstantlyAIConnector(ctx context.Context) *instantlyai.Connector {
	filePath := credscanning.LoadPath(providers.InstantlyAI)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client := utils.NewAPIKeyClient(ctx, reader, providers.InstantlyAI)

	conn, err := instantlyai.NewConnector(common.Parameters{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating InstantlyAI App connector", "error", err)
	}

	return conn
}

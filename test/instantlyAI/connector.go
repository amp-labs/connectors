package instantlyAI

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	instantlyAI "github.com/amp-labs/connectors/providers/instantlyAI"
	"github.com/amp-labs/connectors/test/utils"
)

func GetInstantlyAIConnector(ctx context.Context) *instantlyAI.Connector {
	filePath := credscanning.LoadPath(providers.InstantlyAI)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	info, err := providers.ReadInfo(providers.InstantlyAI)
	if err != nil {
		utils.Fail(err.Error())
	}

	headerName, headerValue, err := info.GetApiKeyHeader(reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail(err.Error())
	}

	client, err := common.NewApiKeyHeaderAuthHTTPClient(
		ctx, headerName, headerValue)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := instantlyAI.NewConnector(common.Parameters{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating InstantlyAI App connector", "error", err)
	}

	return conn
}

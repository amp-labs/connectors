package fireflies

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/fireflies"
	"github.com/amp-labs/connectors/test/utils"
)

func GetFirefliesConnector(ctx context.Context) *fireflies.Connector {
	filePath := credscanning.LoadPath(providers.Fireflies)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	info, err := providers.ReadInfo(providers.Fireflies)
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

	conn, err := fireflies.NewConnector(common.Parameters{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Fireflies App connector", "error", err)
	}

	return conn
}

package heyreach

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/heyreach"
	"github.com/amp-labs/connectors/test/utils"
)

func GetHeyreachConnector(ctx context.Context) *heyreach.Connector {
	filePath := credscanning.LoadPath(providers.HeyReach)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	info, err := providers.ReadInfo(providers.HeyReach)
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

	conn, err := heyreach.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating Heyreach App connector", "error", err)
	}

	return conn
}

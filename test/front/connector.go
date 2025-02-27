package front

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/front"
	"github.com/amp-labs/connectors/test/utils"
	testUtils "github.com/amp-labs/connectors/test/utils"
)

func GetFrontConnector(ctx context.Context) *front.Connector {
	filePath := credscanning.LoadPath(providers.Front)
	reader := testUtils.MustCreateProvCredJSON(filePath, false, false)

	info, err := providers.ReadInfo(providers.Front)
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

	conn, err := front.NewConnector(common.Parameters{
		AuthenticatedClient: client,
	})
	if err != nil {
		testUtils.Fail("error creating Front App connector", "error", err)
	}

	return conn
}

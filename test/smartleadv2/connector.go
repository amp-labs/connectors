package smartlead

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/smartleadv2"
	"github.com/amp-labs/connectors/test/utils"
)

func GetSmartleadV2Connector(ctx context.Context) *smartleadv2.Connector {
	filePath := credscanning.LoadPath(providers.Smartlead)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client, err := common.NewApiKeyQueryParamAuthHTTPClient(ctx, "api_key", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := smartleadv2.NewConnector(
		common.Parameters{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

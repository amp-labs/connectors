package clari

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/clari"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *clari.Connector {
	filePath := credscanning.LoadPath(providers.Clari)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(
		ctx, "apiKey", reader.Get(credscanning.Fields.ApiKey),
	)
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := clari.NewConnector(common.Parameters{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

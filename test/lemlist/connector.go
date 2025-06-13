package lemlist

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/lemlist"
	"github.com/amp-labs/connectors/test/utils"
)

func GetLemlistConnector(ctx context.Context) *lemlist.Connector {
	filePath := credscanning.LoadPath(providers.Lemlist)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyQueryParamAuthHTTPClient(ctx, "access_token", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := lemlist.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

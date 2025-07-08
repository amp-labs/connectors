package flatfile

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/flatfile"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *flatfile.Connector {
	filePath := credscanning.LoadPath(providers.FlatFile)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := flatfile.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

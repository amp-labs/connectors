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
	filePath := credscanning.LoadPath(providers.Flatfile)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	info, err := providers.ReadInfo(providers.Flatfile)
	if err != nil {
		utils.Fail(err.Error())
	}

	headerName, headerValue, err := info.GetApiKeyHeader(reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail(err.Error())
	}

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, headerName, headerValue)
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

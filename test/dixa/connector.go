package dixa

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/dixa"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *dixa.Connector {
	filePath := credscanning.LoadPath(providers.Dixa)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := dixa.NewConnector(
		common.Parameters{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

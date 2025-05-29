package mixmax

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/mixmax"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *mixmax.Connector {
	filePath := credscanning.LoadPath(providers.Mixmax)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "X-API-Token", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := mixmax.NewConnector(parameters.Connector{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

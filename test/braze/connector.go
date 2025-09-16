package braze

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/braze"
	"github.com/amp-labs/connectors/test/utils"
)

func NewBrazeConnector(ctx context.Context) *braze.Connector {
	filePath := credscanning.LoadPath(providers.Braze)

	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", "Bearer "+reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail(err.Error())
	}

	conn, err := braze.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: client,
			Workspace:           "iad-03",
		},
	)
	if err != nil {
		utils.Fail("error creating brevo connector", "error", err)
	}

	return conn
}

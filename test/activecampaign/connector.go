package activecampaign

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/activecampaign"
	"github.com/amp-labs/connectors/test/utils"
)

func GetActiveCampaignConnector(ctx context.Context) *activecampaign.Connector {
	filePath := credscanning.LoadPath(providers.ActiveCampaign)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Api-Token", reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := activecampaign.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: client,
			Workspace:           reader.Get(credscanning.Fields.Workspace),
		},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

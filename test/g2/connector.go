package g2

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/g2"
	"github.com/amp-labs/connectors/test/utils"
)

func NewConnector(ctx context.Context) *g2.Connector {
	filePath := credscanning.LoadPath(providers.G2)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, "Authorization", "Bearer "+reader.Get(credscanning.Fields.ApiKey))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := g2.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client, Metadata: map[string]string{
			"subject_product_id": "jira",
		}},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

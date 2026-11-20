package twenty

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/twenty"
	"github.com/amp-labs/connectors/test/utils"
)

// GetTwentyConnector reads twenty-creds.json (or the file named by TWENTY_CRED_FILE):
//
//	{"provider": "twenty", "apiKey": "<workspace API key>", "metadata": {"workspace": "crm.yourcompany.com"}}
func GetTwentyConnector(ctx context.Context) *twenty.Connector {
	filePath := credscanning.LoadPath(providers.Twenty)
	reader := utils.MustCreateProvCredJSON(filePath, false, credscanning.Fields.Workspace)

	conn, err := twenty.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Twenty),
			Workspace:           reader.Get(credscanning.Fields.Workspace),
		},
	)
	if err != nil {
		utils.Fail("error creating Twenty connector", "error", err)
	}

	return conn
}

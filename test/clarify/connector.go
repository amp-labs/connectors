package clarify

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/clarify"
	"github.com/amp-labs/connectors/test/utils"
)

func GetClarifyConnector(ctx context.Context) *clarify.Connector {
	filePath := credscanning.LoadPath(providers.Clarify)

	fieldWorkspace := credscanning.Field{Name: "workspace", PathJSON: "metadata.workspace"}
	reader := utils.MustCreateProvCredJSON(filePath, false, fieldWorkspace)

	conn, err := clarify.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Clarify),
			Workspace:           reader.Get(fieldWorkspace),
		},
	)
	if err != nil {
		utils.Fail("error creating Clarify connector", "error", err)
	}

	return conn
}

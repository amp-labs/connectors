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

	reader := utils.MustCreateProvCredJSON(filePath, false,
		credscanning.Field{Name: "workspace", PathJSON: "metadata.workspace"},
	)

	conn, err := clarify.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Clarify),
			Metadata: map[string]string{
				"workspace": reader.Get(credscanning.Field{Name: "workspace"}),
			},
		},
	)
	if err != nil {
		utils.Fail("error creating Clarify connector", "error", err)
	}

	return conn
}

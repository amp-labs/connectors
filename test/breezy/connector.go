package breezy

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/breezy"
	"github.com/amp-labs/connectors/test/utils"
)

func GetBreezyConnector(ctx context.Context) *breezy.Connector {
	filePath := credscanning.LoadPath(providers.Breezy)

	reader := utils.MustCreateProvCredJSON(filePath, false,
		credscanning.Field{Name: "company_id", PathJSON: "metadata.company_id"},
	)

	conn, err := breezy.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Breezy),
			Metadata: map[string]string{
				"company_id": reader.Get(credscanning.Field{Name: "company_id"}),
			},
		},
	)
	if err != nil {
		utils.Fail("error creating Breezy connector", "error", err)
	}

	return conn
}

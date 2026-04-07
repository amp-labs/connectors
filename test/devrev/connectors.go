package devrev

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/devrev"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *devrev.Connector {
	filePath := credscanning.LoadPath(providers.DevRev)
	reader := utils.MustCreateProvCredJSON(filePath, false, credscanning.Fields.Token)
	token := reader.Get(credscanning.Fields.Token)
	client, err := common.NewCustomAuthHTTPClient(ctx,
		common.WithCustomHeaders(common.Header{Key: "Authorization", Value: "Bearer " + token}),
	)
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := devrev.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

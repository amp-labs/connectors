package connectwise

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/connectwise"
	"github.com/amp-labs/connectors/test/utils"
)

var clientIdField = credscanning.Field{
	Name:      "clientId",
	PathJSON:  "metadata.clientId",
	SuffixENV: "CLIENT_ID",
}

var regionField = credscanning.Field{
	Name:      "region",
	PathJSON:  "metadata.region",
	SuffixENV: "REGION",
}

func GetConnectWiseConnector(ctx context.Context) *connectwise.Connector {
	filePath := credscanning.LoadPath(providers.ConnectWise)
	reader := utils.MustCreateProvCredJSON(filePath, false, clientIdField, regionField)

	conn, err := connectwise.NewConnector(common.ConnectorParams{
		AuthenticatedClient: utils.NewBasicAuthClient(ctx, reader),
		Metadata: map[string]string{
			"clientId": reader.Get(clientIdField),
			"region":   reader.Get(regionField),
		},
	})
	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

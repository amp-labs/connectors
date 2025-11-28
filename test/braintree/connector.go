package braintree

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/braintree"
	"github.com/amp-labs/connectors/test/utils"
)

func GetBraintreeConnector(ctx context.Context) *braintree.Connector {
	filePath := credscanning.LoadPath(providers.Braintree)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	client := utils.NewBasicAuthClient(ctx, reader)

	conn, err := braintree.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
	})
	if err != nil {
		utils.Fail("error creating braintree connector", "error", err)
	}

	return conn
}

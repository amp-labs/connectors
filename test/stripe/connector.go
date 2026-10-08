package stripe

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/stripe"
	"github.com/amp-labs/connectors/test/utils"
)

func GetStripeConnector(ctx context.Context) *stripe.Connector {
	filePath := credscanning.LoadPath(providers.Stripe)
	reader := utils.MustCreateProvCredJSON(filePath, false)

	conn, err := stripe.NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: utils.NewAPIKeyClient(ctx, reader, providers.Stripe),
		},
	)
	if err != nil {
		utils.Fail("error creating Stripe connector", "error", err)
	}

	return conn
}

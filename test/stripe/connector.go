package stripe

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/stripe"
	"github.com/amp-labs/connectors/test/utils"
)

func GetStripeConnector(ctx context.Context) *stripe.Connector {
	filePath := credscanning.LoadPath(providers.Stripe)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := stripe.NewConnector(
		stripe.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		utils.Fail("error creating Stripe connector", "error", err)
	}

	return conn
}

package journeysapp

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/customerapp"
	"github.com/amp-labs/connectors/test/utils"
)

func GetCustomerJourneysAppConnector(ctx context.Context) *customerapp.Connector {
	filePath := credscanning.LoadPath(providers.CustomerJourneysApp)
	reader := utils.MustCreateProvCredJSON(filePath, false, false)

	conn, err := customerapp.NewConnector(
		customerapp.WithClient(ctx, http.DefaultClient,
			reader.Get(credscanning.Fields.ApiKey),
		),
	)
	if err != nil {
		utils.Fail("error creating Customer Journeys App connector", "error", err)
	}

	return conn
}

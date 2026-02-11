package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "Account",
		Fields:     connectors.Fields("Id", "Name", "BillingCity", "IsDeleted", "SystemModstamp"),
		Filter: connectors.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "Name",
					Operator:  common.FilterOperatorEQ,
					Value:     "OKADA TEST ACCOUNT",
				},
			},
		},
	})
}

package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("email", "phone", "company", "website", "lastname", "firstname"),
		Filter: connectors.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "firstname",
					Operator:  common.FilterOperatorEQ,
					Value:     "Johnnie",
				},
			},
		},
		Limit: 50,
		AssociatedObjects: []string{
			"companies",
		},
	})
}

package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/connectwise"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnectWiseConnector(ctx)

	testscenario.SearchThroughPages(ctx, conn, connectors.SearchParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("customfield16", "firstname", "lastname"),
		Filter: connectors.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "id",
					Operator:  common.FilterOperatorIn,
					Value:     []string{"15", "16", "17"},
				},
			},
		},
		Limit: 2,
	})
}

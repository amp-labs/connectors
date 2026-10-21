package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/orb"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	reads := []common.ReadParams{
		{ObjectName: "customers", Fields: connectors.Fields("id", "external_customer_id", "name", "timezone")},
		{ObjectName: "subscriptions", Fields: connectors.Fields("id", "status", "customer", "start_date")},
		{ObjectName: "invoices", Fields: connectors.Fields("id", "status", "total", "invoice_date")},
		{ObjectName: "metrics", Fields: connectors.Fields("id", "name", "status")},
	}

	for _, params := range reads {
		params.PageSize = 3
		testscenario.ReadThroughPages(ctx, conn, params)
	}
}

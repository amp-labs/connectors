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

	objects := map[string][]string{
		"customers":     {"id", "external_customer_id", "name", "timezone"},
		"subscriptions": {"id", "customer", "status", "start_date"},
		"invoices":      {"id", "customer", "status", "total"},
		"metrics":       {"id", "name", "status"},
	}

	for objectName, fields := range objects {
		testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields(fields...),
			PageSize:   2,
		})
	}
}

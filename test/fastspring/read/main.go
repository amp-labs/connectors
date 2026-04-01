package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/fastspring"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetFastSpringConnector(ctx)

	slog.Info("=== Reading accounts ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("id", "account", "language", "country"),
		PageSize:   50,
	})

	slog.Info("=== Reading orders ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "orders",
		Fields:     connectors.Fields("order", "id", "reference"),
		PageSize:   50,
	})

	slog.Info("=== Reading products ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("path"),
		PageSize:   50,
	})

	slog.Info("=== Reading subscriptions ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "subscriptions",
		Fields:     connectors.Fields("subscription", "id", "product", "state", "currency"),
		PageSize:   50,
	})

	slog.Info("=== Reading unprocessed events ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "events-unprocessed",
		Fields:     connectors.Fields("id", "event", "type", "processed", "live", "created"),
		PageSize:   50,
	})

	slog.Info("=== Reading processed events ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "events-processed",
		Fields:     connectors.Fields("id", "event", "type", "processed", "live", "created"),
		PageSize:   50,
	})
}

package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/shopify"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := shopify.GetShopifyConnector(ctx)

	slog.Info("=== Reading all customers ===")

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("id", "firstName", "lastName", "email"),
	})
}
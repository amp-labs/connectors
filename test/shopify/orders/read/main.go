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

	utils.SetupLogging()
	conn := shopify.GetShopifyConnector(ctx)

	slog.Info("=== Reading all orders ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "orders",
		Fields:     connectors.Fields("id", "name", "email", "updatedAt"),
	})
}

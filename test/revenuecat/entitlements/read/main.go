package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetRevenueCatConnector(ctx)

	slog.Info("=== Reading all entitlements ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "entitlements",
		Fields:     connectors.Fields("id", "lookup_key", "display_name"),
	})
}

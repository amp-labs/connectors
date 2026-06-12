package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/breezy"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetBreezyConnector(ctx)

	slog.Info("=== Read companies ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "companies",
		Fields:     connectors.Fields("_id", "name"),
	})

	slog.Info("=== Read positions ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "positions",
		Fields:     connectors.Fields("_id", "name", "state"),
	})

	slog.Info("=== Read webhook endpoints ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "webhook_endpoints",
		Fields:     connectors.Fields("id", "url", "status"),
	})

	slog.Info("=== Metadata vs Read validation ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "companies", nil)
	testscenario.ValidateMetadataContainsRead(ctx, conn, "positions", nil)
	testscenario.ValidateMetadataContainsRead(ctx, conn, "webhook_endpoints", nil)

	slog.Info("Breezy read tests completed successfully!")
}

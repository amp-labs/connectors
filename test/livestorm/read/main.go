package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/livestorm"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetLivestormConnector(ctx)

	slog.Info("=== Basic paginated read: events ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "events",
		Fields: connectors.Fields(
			"title",
			"updated_at",
		),
	})

	slog.Info("=== Incremental read with Since: events ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "events",
		Fields: connectors.Fields(
			"title",
			"updated_at",
		),
		Since: time.Now().UTC().AddDate(0, -1, 0),
	})

	slog.Info("=== Basic paginated read: people ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "people",
		Fields: connectors.Fields(
			"email",
			"first_name",
			"last_name",
		),
	})

	slog.Info("=== Basic paginated read: people_attributes ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "people_attributes",
		Fields: connectors.Fields(
			"slug",
			"name",
			"type",
		),
	})

	slog.Info("=== Metadata vs Read validation: events ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "events", nil)

	slog.Info("=== Metadata vs Read validation: people ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "people", nil)

	slog.Info("=== Metadata vs Read validation: people_attributes ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "people_attributes", nil)

	slog.Info("Livestorm read tests completed successfully!")
}

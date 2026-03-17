package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	slog.Info("=== Basic paginated read: campaigns ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "description", "createdOn", "isDefault"),
		PageSize:   1,
	})

	slog.Info("=== Custom filter read: campaigns with isDefault=true sorted by createdOn DESC ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "createdOn"),
		Filter:     "query[isDefault]=true&sort[createdOn]=DESC",
		PageSize:   1,
	})

	slog.Info("=== Incremental read with Since: campaigns (connector-side filtering) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "createdOn"),
		Since:      time.Now().AddDate(0, -6, 0), // last 6 months
	})

	slog.Info("=== Metadata vs Read validation: campaigns ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "campaigns", nil)

	slog.Info("Campaigns read tests completed successfully!")
}

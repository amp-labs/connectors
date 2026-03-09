package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
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

	slog.Info("=== Basic paginated read: contacts ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     datautils.NewSet("contactId", "email", "name", "createdOn"),
		PageSize:   1,
	})

	slog.Info("=== Custom filter read: campaigns with isDefault=true sorted by createdOn DESC ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "createdOn"),
		Filter:     "query[isDefault]=true&sort[createdOn]=DESC",
		PageSize:   1,
	})

	slog.Info("=== Incremental read with Since: contacts (provider-side filtering) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contactId", "email", "name", "createdOn"),
		Since:      time.Now().AddDate(0, -1, 0), // last 30 days
	})

	slog.Info("=== Incremental read with Since: campaigns (connector-side filtering) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("campaignId", "name", "createdOn"),
		Since:      time.Now().AddDate(0, -6, 0), // last 6 months
	})

	slog.Info("=== Basic read: newsletters ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "newsletters",
		Fields:     connectors.Fields("newsletterId", "name", "createdOn"),
	})

	slog.Info("=== Basic read: autoresponders ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "autoresponders",
		Fields:     connectors.Fields("autoresponderId", "name", "createdOn"),
	})

	// Forms require account/campaign permission; collaborators or restricted accounts may get 403.
	// Uncomment to test when the connected account has form access.
	// slog.Info("=== Basic read: forms ===")
	// testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
	// 	ObjectName: "forms",
	// 	Fields:     connectors.Fields("formId", "name", "createdOn"),
	// })

	slog.Info("=== Basic read: custom-events ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "custom-events",
		Fields:     connectors.Fields("customEventId", "name", "createdOn"),
	})

	slog.Info("=== Basic read: webinars ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "webinars",
		Fields:     connectors.Fields("webinarId", "name", "createdOn"),
	})

	slog.Info("=== Basic read: landing-pages ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "landing-pages",
		Fields:     connectors.Fields("landingPageId", "name", "createdOn"),
	})

	slog.Info("=== Metadata vs Read validation: contacts ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "contacts", nil)

	slog.Info("=== Metadata vs Read validation: campaigns ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "campaigns", nil)

	slog.Info("All read tests completed successfully!")
}

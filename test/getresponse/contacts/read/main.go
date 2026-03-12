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

	slog.Info("=== Basic paginated read: contacts ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contactId", "email", "name", "createdOn"),
		PageSize:   1,
	})

	slog.Info("=== Incremental read with Since: contacts (provider-side filtering) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contactId", "email", "name", "createdOn"),
		Since:      time.Now().AddDate(0, -1, 0), // last 30 days
	})

	slog.Info("=== Metadata vs Read validation: contacts ===")
	testscenario.ValidateMetadataContainsRead(ctx, conn, "contacts", nil)

	slog.Info("Contacts read tests completed successfully!")
}

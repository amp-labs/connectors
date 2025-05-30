package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "Emails" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceAccountEngagementConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	slog.Info("Metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}

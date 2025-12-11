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

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"TestObject15__c",
	})
	if err != nil {
		utils.Fail("error listing metadata for Salesforce", "error", err)
	}

	slog.Info("Metadata for accounts..")
	utils.DumpJSON(metadata, os.Stdout)
}

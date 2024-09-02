package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/salesforce/bulk"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	defer utils.Close(conn)

	query := "SELECT Id, Name FROM Account LIMIT 1000"

	res, err := conn.BulkQuery(ctx, query, false)
	if err != nil {
		utils.Fail("Error querying", "error", err)
	}

	slog.Info("Query Info")
	utils.DumpJSON(res, os.Stdout)

	bulk.LoadQueryResults(ctx, conn, res.Id)
}

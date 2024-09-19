package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
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

	res, err := conn.BulkRead(ctx, common.ReadParams{
		ObjectName: "Account",
		Fields:     connectors.Fields("Id", "Name"),
	})
	if err != nil {
		utils.Fail("error bulk reading from Salesforce", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)

	bulk.LoadQueryResults(ctx, conn, res.Id)
}

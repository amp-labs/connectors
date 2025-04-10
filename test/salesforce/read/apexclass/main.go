package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
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

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "ApexClass",
		Fields:     connectors.Fields("*"),
	})
	if err != nil {
		utils.Fail("error reading from Salesforce: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)
}

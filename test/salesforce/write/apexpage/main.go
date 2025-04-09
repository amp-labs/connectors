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

	res, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "ApexPage",
		RecordData: map[string]any{
			"ApiVersion":  57.0,
			"Name":        "SimplePage",
			"MasterLabel": "Create",
			"Markup":      "<apex:page><h1>Hello</h1></apex:page>",
		},
	})
	if err != nil {
		utils.Fail("error reading from Salesforce: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)
}

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
		ObjectName: "User",
		RecordId:   "005ak00000FthwYAAR",
		RecordData: map[string]any{
			//"Username":          "teddy.anderson2@example.com",
			//"LastName":          "Teddy2",
			//"Alias":             "Ted2",
			//"Email":             "newuser@example.com",
			//"TimeZoneSidKey":    "America/Los_Angeles",
			//"LocaleSidKey":      "en_US",
			//"EmailEncodingKey":  "UTF-8",
			//"LanguageLocaleKey": "en_US",
			//"ProfileId":         "00eak00000CDaoLAAT",
			"UserRoleId": "00Eak000004uhBNEAY",
		},
	})
	if err != nil {
		utils.Fail("error reading from Salesforce: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)
}

package main

import (
	"context"
	"fmt"
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

	res, err := conn.BatchWrite(ctx, &connectors.BatchWriteParam{
		ObjectName: "Contact",
		Type:       connectors.BatchWriteTypeUpdate,
		Records: []any{
			map[string]any{
				"id":        "003ak00000jvIfpAAE",
				"LastName":  "Dyer (updated)",
				"FirstName": "Siena (updated)",
			},
			map[string]any{
				"id":        "003ak00000jvIfqAAE",
				"LastName":  "Blevins (updated)",
				"FirstName": "Markus (updated)",
			},
		},
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	fmt.Println("Reading..")
	utils.DumpJSON(res, os.Stdout)
}

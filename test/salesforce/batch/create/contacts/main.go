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
		Type:       connectors.BatchWriteTypeCreate,
		Records: []any{
			map[string]any{
				"LastName":  "Dyer",
				"FirstName": "Siena",
			},
			map[string]any{
				"LastName":  "Blevins",
				"FirstName": "Markus",
			},
		},
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	fmt.Println("Reading..")
	utils.DumpJSON(res, os.Stdout)
}

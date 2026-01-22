package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
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
		Type:       connectors.WriteTypeCreate,
		Batch: common.BatchItems{{
			Record: map[string]any{
				"LastName":  "Dyer",
				"FirstName": "Siena",
			},
		}, {
			Record: map[string]any{
				"LastName":  "Blevins",
				"FirstName": "Markus",
			},
		}},
	})
	if err != nil {
		utils.Fail("error reading", "error", err)
	}

	fmt.Println("Creating..")
	utils.DumpJSON(res, os.Stdout)
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Turn on verbose logging
	ctx = logging.WithLoggerEnabled(ctx, true)
	ctx = logging.WithVerboseLogging(ctx, true)

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "marketing-campaigns",
		Fields:     connectors.Fields("hs_name", "hs_notes", "hs_budget_items_sum_amount"),
		AssociatedObjects: []string{
			//"assets",
			"contacts",
		},
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	fmt.Println("Reading...") //nolint:forbidigo

	// Dump the result.
	utils.DumpJSON(res, os.Stdout)
}

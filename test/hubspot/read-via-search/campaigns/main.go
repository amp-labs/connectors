package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers/hubspot"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetHubspotConnector(ctx)

	res, err := conn.ReadUsingSearchAPI(ctx, hubspot.SearchParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("hs_name", "hs_notes", "hs_budget_items_sum_amount"),
		Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)
}

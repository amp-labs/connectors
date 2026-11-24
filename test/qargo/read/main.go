package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/qargo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := qargo.NewConnector(ctx)

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "company",
		Since:      time.Now().Add(-30 * 24 * time.Hour),
		Fields: connectors.Fields("id", "name", "code", "is_customer", "payment_term",
			"billing_location", "timestamp_updated"),
	})
	if err != nil {
		utils.Fail("error reading companies", "error", err)
	}

	fmt.Println("Reading companies")
	utils.DumpJSON(res, os.Stdout)

	// Invoices and credit notes waiting in the accounting sync-task queue. Reading does not
	// complete their tasks, so they come back on every run.
	for _, objectName := range []string{"sales-invoice", "sales-credit-note"} {
		res, err = conn.Read(ctx, connectors.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields("id", "name", "date", "customer", "billing_entity", "line_items"),
			PageSize:   5,
		})
		if err != nil {
			utils.Fail("error reading "+objectName, "error", err)
		}

		fmt.Println("Reading " + objectName)
		utils.DumpJSON(res, os.Stdout)
	}
}

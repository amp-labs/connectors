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
}

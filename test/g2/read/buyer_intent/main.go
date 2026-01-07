package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/g2"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := g2.NewConnector(ctx)

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "buyer_intent",
		Since:      time.Now().Add(-3000 * time.Hour),
		Fields:     connectors.Fields("signal_type", "time", "day", "week", "company_domain", "company_country"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, connectors.ReadParams{
		ObjectName: "products",
		Since:      time.Now().Add(-3000 * time.Hour),
		Fields:     connectors.Fields("detail_description", "domain", "g2_url", "image_url", "name", "public_detail_url"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)
}

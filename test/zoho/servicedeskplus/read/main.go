package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx, providers.ModuleZohoServiceDeskPlus)

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "requests",
		Since:      time.Now().Add(-3000 * time.Hour),
		Fields:     connectors.Fields("id", "created_time", "has_draft"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, connectors.ReadParams{
		ObjectName: "assets",
		Since:      time.Now().Add(-10 * time.Hour),
		Fields:     connectors.Fields("id", "purchase_cost", "last_updated_time", "created_time"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)
}

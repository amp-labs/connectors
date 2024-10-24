package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/zohocrm"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := zohocrm.GetZohoConnector(ctx)
	defer utils.Close(conn)

	m, err := conn.ListObjectMetadata(ctx, []string{"leads", "deals", "contacts"})
	if err != nil {
		utils.Fail("error listing metadata for Zoho", "error", err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}

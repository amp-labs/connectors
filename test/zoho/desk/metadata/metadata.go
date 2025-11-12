package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := zoho.GetZohoConnector(ctx, providers.ModuleZohoDesk)

	m, err := conn.ListObjectMetadata(ctx, []string{"accounts", "contacts", "agents"})
	if err != nil {
		utils.Fail("error listing metadata for Zoho", "error", err)
	}

	// Print the results
	utils.DumpJSON(m.Result, os.Stdout)
	fmt.Println("Errors: ", m.Errors)
}

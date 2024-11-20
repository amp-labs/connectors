package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/pipedrive"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := pipedrive.GetPipedriveConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"activities", "callLogs", "currencies", "deals"})
	if err != nil {
		utils.Fail("error listing metadata for Pipedrive", "error", err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}

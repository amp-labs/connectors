package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/pipedrive"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := pipedrive.GetPipedriveConnector(ctx, providers.ModulePipedriveCRM)

	m, err := conn.ListObjectMetadata(ctx, []string{"activities", "stages", "deals"})
	if err != nil {
		utils.Fail("error listing metadata for Pipedrive", "error", err)
	}

	// Print the results
	utils.DumpJSON(m.Result, os.Stdout)
	fmt.Println("Errors: ", m.Errors)
}

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesflare"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesflareConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"me/contacts",
		"opportunities",
	})
	if err != nil {
		utils.Fail("error listing metadata for microsoft CRM", "error", err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}

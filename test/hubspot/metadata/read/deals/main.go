package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"deals",
	})
	if err != nil {
		utils.Fail("error listing metadata for Hubspot", "error", err)
	}

	slog.Info("Metadata deals..")
	utils.DumpJSON(metadata, os.Stdout)
}

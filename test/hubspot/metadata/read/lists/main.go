package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "lists"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Hubspot", "error", err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}

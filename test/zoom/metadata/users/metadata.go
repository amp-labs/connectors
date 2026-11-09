package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoom"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZoomConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{"users"})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}

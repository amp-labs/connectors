package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/copper"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCopperConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"people", "tags",
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	fmt.Println("Metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}

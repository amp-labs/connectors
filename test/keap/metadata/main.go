package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/keap"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "contacts"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetKeapConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Keap", "error", err)
	}

	fmt.Println("Metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}

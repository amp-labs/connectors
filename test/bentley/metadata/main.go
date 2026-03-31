package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/bentley"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetBentleyConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"contextcapture/jobs",
		"itwins/favorites",
		"library/manufacturers",
		"webhooks",
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	for objName, objMeta := range metadata.Result {
		fmt.Printf("  %s (%d fields)\n", objName, len(objMeta.Fields))
	}

	if len(metadata.Errors) > 0 {
		fmt.Println("Errors:")
		for obj, err := range metadata.Errors {
			fmt.Printf("  %s: %v\n", obj, err)
		}
	}

	utils.DumpJSON(metadata, os.Stdout)
}

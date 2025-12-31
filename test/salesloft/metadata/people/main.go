package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesloft"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSalesloftConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"people",
	})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	utils.DumpJSON(metadata, os.Stdout)
}

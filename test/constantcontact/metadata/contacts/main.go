package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/constantcontact"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConstantContactConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		"contacts",
	})
	if err != nil {
		utils.Fail("error listing metadata for ConstantContact", "error", err)
	}

	slog.Info("Metadata contacts..")
	utils.DumpJSON(metadata, os.Stdout)
}

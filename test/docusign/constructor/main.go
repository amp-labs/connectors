package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/docusign"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	connTest.GetDocusignConnector(ctx)

	slog.Info("constructor finished")
}

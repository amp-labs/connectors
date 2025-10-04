package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/test/calendly"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := calendly.GetConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"scheduled_events"})
	if err != nil {
		utils.Fail("error listing metadata for Calendly", "error", err)
	}

	slog.Info("Metadata for scheduled_events")
	utils.DumpJSON(m, os.Stdout)
} 